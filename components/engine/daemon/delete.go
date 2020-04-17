package daemon // import "github.com/docker/docker/daemon"

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/volume"
	volumestore "github.com/docker/docker/volume/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContainerRm removes the container id from the filesystem. An error
// is returned if the container is not found, or if the remove
// fails. If the remove succeeds, the container name is released, and
// network links are removed.
func (daemon *Daemon) ContainerRm(name string, config *types.ContainerRmConfig) error {
	// 得到对应container对象
	start := time.Now()
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}

	// 设置container状态为 RemovalInProgress (一种中间状态)
	// 为了避免同时调用docker命令的竞争
	// 在rm结束时会取消 RemovalInProgress 状态
	// Container state RemovalInProgress should be used to avoid races.
	if inProgress := container.SetRemovalInProgress(); inProgress {
		err := fmt.Errorf("removal of container %s is already in progress", name)
		return errdefs.Conflict(err)
	}
	defer container.ResetRemovalInProgress()

	// 再次得到对应container对象（两次检查）
	// check if container wasn't deregistered by previous rm since Get
	if c := daemon.containers.Get(container.ID); c == nil {
		return nil
	}

	// rm指定删除link
	if config.RemoveLink {
		return daemon.rmLink(container, name)
	}

	// 停止容器, 清理相关资源
	err = daemon.cleanupContainer(container, config.ForceRemove, config.RemoveVolume)
	containerActions.WithValues("delete").UpdateSince(start)

	return err
}

func (daemon *Daemon) rmLink(container *container.Container, name string) error {
	if name[0] != '/' {
		name = "/" + name
	}
	parent, n := path.Split(name)
	if parent == "/" {
		return fmt.Errorf("Conflict, cannot remove the default name of the container")
	}

	parent = strings.TrimSuffix(parent, "/")
	pe, err := daemon.containersReplica.Snapshot().GetID(parent)
	if err != nil {
		return fmt.Errorf("Cannot get parent %s for name %s", parent, name)
	}

	daemon.releaseName(name)
	parentContainer, _ := daemon.GetContainer(pe)
	if parentContainer != nil {
		daemon.linkIndex.unlink(name, container, parentContainer)
		if err := daemon.updateNetwork(parentContainer); err != nil {
			logrus.Debugf("Could not update network to remove link %s: %v", n, err)
		}
	}
	return nil
}

// cleanupContainer unregisters a container from the daemon, stops stats
// collection and cleanly removes contents and metadata from the filesystem.
func (daemon *Daemon) cleanupContainer(container *container.Container, forceRemove, removeVolume bool) (err error) {
	// 如果容器是running状态, 并且指定了forceRemove, 那么停止容器(直接使用SIGKILL)
	if container.IsRunning() {
		if !forceRemove {
			state := container.StateString()
			procedure := "Stop the container before attempting removal or force remove"
			if state == "paused" {
				procedure = "Unpause and then " + strings.ToLower(procedure)
			}
			err := fmt.Errorf("You cannot remove a %s container %s. %s", state, container.ID, procedure)
			return errdefs.Conflict(err)
		}
		if err := daemon.Kill(container); err != nil {
			return fmt.Errorf("Could not kill running container %s, cannot remove - %v", container.ID, err)
		}
	}
	// 检查os
	if !system.IsOSSupported(container.OS) {
		return fmt.Errorf("cannot remove %s: %s ", container.ID, system.ErrNotSupportedOperatingSystem)
	}

	// 停止正在进行stats推送的其他goroutine
	// stop collection of stats for the container regardless
	// if stats are currently getting collected.
	daemon.statsCollector.StopCollection(container)

	// 进行容器的停止
	// 上面处理了running状态的, 所以其他状态都会经过这里
	if err = daemon.containerStop(container, 3); err != nil {
		return err
	}

	// 加锁container的单独锁, 进行checkpoint(记录state至磁盘)
	// Mark container dead. We don't want anybody to be restarting it.
	container.Lock()
	container.Dead = true

	// Save container state to disk. So that if error happens before
	// container meta file got removed from disk, then a restart of
	// docker should not make a dead container alive.
	if err := container.CheckpointTo(daemon.containersReplica); err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Error saving dying container to disk: %v", err)
	}
	container.Unlock()

	// 通过 ImageService.ReleaseLayer() 删除对应的rwlayer
	// When container creation fails and `RWLayer` has not been created yet, we
	// do not call `ReleaseRWLayer`
	if container.RWLayer != nil {
		err := daemon.imageService.ReleaseLayer(container.RWLayer, container.OS)
		if err != nil {
			err = errors.Wrapf(err, "container %s", container.ID)
			container.SetRemovalError(err)
			return err
		}
		container.RWLayer = nil
	}

	// 删除container对应的Root目录(/var/lib/docker/containers/[id])
	if err := system.EnsureRemoveAll(container.Root); err != nil {
		e := errors.Wrapf(err, "unable to remove filesystem for %s", container.ID)
		container.SetRemovalError(e)
		return e
	}

	// 各个store中删除对应记录
	linkNames := daemon.linkIndex.delete(container)
	selinuxFreeLxcContexts(container.ProcessLabel)
	daemon.idIndex.Delete(container.ID)
	daemon.containers.Delete(container.ID)
	daemon.containersReplica.Delete(container)
	if e := daemon.removeMountPoints(container, removeVolume); e != nil {
		logrus.Error(e)
	}
	for _, name := range linkNames {
		daemon.releaseName(name)
	}
	container.SetRemoved()
	stateCtr.del(container.ID)

	daemon.LogContainerEvent(container, "destroy")
	return nil
}

// VolumeRm removes the volume with the given name.
// If the volume is referenced by a container it is not removed
// This is called directly from the Engine API
func (daemon *Daemon) VolumeRm(name string, force bool) error {
	v, err := daemon.volumes.Get(name)
	if err != nil {
		if force && volumestore.IsNotExist(err) {
			return nil
		}
		return err
	}

	err = daemon.volumeRm(v)
	if err != nil && volumestore.IsInUse(err) {
		return errdefs.Conflict(err)
	}

	if err == nil || force {
		daemon.volumes.Purge(name)
		return nil
	}
	return err
}

func (daemon *Daemon) volumeRm(v volume.Volume) error {
	if err := daemon.volumes.Remove(v); err != nil {
		return errors.Wrap(err, "unable to remove volume")
	}
	daemon.LogVolumeEvent(v.Name(), "destroy", map[string]string{"driver": v.DriverName()})
	return nil
}
