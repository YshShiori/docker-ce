package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"runtime"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/mount"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContainerStart starts a container.
func (daemon *Daemon) ContainerStart(name string, hostConfig *containertypes.HostConfig,
	checkpoint string, checkpointDir string) error {
	if checkpoint != "" && !daemon.HasExperimental() {
		return errdefs.InvalidParameter(errors.New("checkpoint is only supported in experimental mode"))
	}

	// 从 ContainerStore 与 TruncIndex 中得到name对应的container对象
	container, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}

	// 检查State合法
	// Pasued、Running、RemovalInProgress、Dead 状态不能调用start
	validateState := func() error {
		container.Lock()
		defer container.Unlock()

		if container.Paused {
			return errdefs.Conflict(errors.New("cannot start a paused container, try unpause instead"))
		}

		if container.Running {
			return containerNotModifiedError{running: true}
		}

		if container.RemovalInProgress || container.Dead {
			return errdefs.Conflict(errors.New("container is marked for removal and cannot be started"))
		}
		return nil
	}

	if err := validateState(); err != nil {
		return err
	}

	// container的hostconfig配置
	// Windows does not have the backwards compatibility issue here.
	if runtime.GOOS != "windows" {
		// This is kept for backward compatibility - hostconfig should be passed when
		// creating a container, not during start.
		if hostConfig != nil {
			logrus.Warn("DEPRECATED: Setting host configuration options when the container starts is deprecated and has been removed in Docker 1.12")
			oldNetworkMode := container.HostConfig.NetworkMode
			if err := daemon.setSecurityOptions(container, hostConfig); err != nil {
				return errdefs.InvalidParameter(err)
			}
			if err := daemon.mergeAndVerifyLogConfig(&hostConfig.LogConfig); err != nil {
				return errdefs.InvalidParameter(err)
			}
			if err := daemon.setHostConfig(container, hostConfig); err != nil {
				return errdefs.InvalidParameter(err)
			}
			newNetworkMode := container.HostConfig.NetworkMode
			if string(oldNetworkMode) != string(newNetworkMode) {
				// if user has change the network mode on starting, clean up the
				// old networks. It is a deprecated feature and has been removed in Docker 1.12
				container.NetworkSettings.Networks = nil
				if err := container.CheckpointTo(daemon.containersReplica); err != nil {
					return errdefs.System(err)
				}
			}
			container.InitDNSHostConfig()
		}
	} else {
		if hostConfig != nil {
			return errdefs.InvalidParameter(errors.New("Supplying a hostconfig on start is not supported. It should be supplied on create"))
		}
	}

	// 检查hostconfig
	// check if hostConfig is in line with the current system settings.
	// It may happen cgroups are umounted or the like.
	if _, err = daemon.verifyContainerSettings(container.OS, container.HostConfig, nil, false); err != nil {
		return errdefs.InvalidParameter(err)
	}
	// Adapt for old containers in case we have updates in this function and
	// old containers never have chance to call the new function in create stage.
	if hostConfig != nil {
		if err := daemon.adaptContainerSettings(container.HostConfig, false); err != nil {
			return errdefs.InvalidParameter(err)
		}
	}

	// 调用 containeStart 真实开始
	return daemon.containerStart(container, checkpoint, checkpointDir, true)
}

// containerStart prepares the container to run by setting up everything the
// container needs, such as storage and networking, as well as links
// between containers. The container is left waiting for a signal to
// begin running.
func (daemon *Daemon) containerStart(container *container.Container,
	checkpoint string, checkpointDir string, resetRestartManager bool) (err error) {
	start := time.Now()
	container.Lock()
	defer container.Unlock()

	// 保证状态
	if resetRestartManager && container.Running { // skip this check if already in restarting step and resetRestartManager==false
		return nil
	}

	if container.RemovalInProgress || container.Dead {
		return errdefs.Conflict(errors.New("container is marked for removal and cannot be started"))
	}

	if checkpointDir != "" {
		// TODO(mlaventure): how would we support that?
		return errdefs.Forbidden(errors.New("custom checkpointdir is not supported"))
	}

	// 回滚操作, 如果启动失败, 将失败信息记录到container的部分属性(Error, ExitError), 并且清理所有相关的资源
	// 如果设置的--rm, 那么同时删除容器信息
	// if we encounter an error during start we need to ensure that any other
	// setup has been cleaned up properly
	defer func() {
		if err != nil {
			container.SetError(err)
			// if no one else has set it, make sure we don't leave it at zero
			if container.ExitCode() == 0 {
				container.SetExitCode(128)
			}
			if err := container.CheckpointTo(daemon.containersReplica); err != nil {
				logrus.Errorf("%s: failed saving state on start failure: %v", container.ID, err)
			}
			container.Reset(false)

			// 清理容器相关资源
			daemon.Cleanup(container)
			// 自动remove container
			// if containers AutoRemove flag is set, remove it after clean up
			if container.HostConfig.AutoRemove {
				container.Unlock()
				if err := daemon.ContainerRm(container.ID, &types.ContainerRmConfig{ForceRemove: true, RemoveVolume: true}); err != nil {
					logrus.Errorf("can't remove container %s: %v", container.ID, err)
				}
				container.Lock()
			}
		}
	}()

	// 进行容器的rwlayer的union mount(依赖于graphdriver实现), 并将rootfs路径保存到 container.BaseFs属性中
	if err := daemon.conditionalMountOnStart(container); err != nil {
		return err
	}

	// 容器网络资源的准备
	if err := daemon.initializeNetworking(container); err != nil {
		return err
	}

	// 容器配置转变为spec
	spec, err := daemon.createSpec(container)
	if err != nil {
		return errdefs.System(err)
	}

	// restart count置为0
	if resetRestartManager {
		container.ResetRestartManager(true)
	}

	if daemon.saveApparmorConfig(container); err != nil {
		return err
	}

	if checkpoint != "" {
		checkpointDir, err = getCheckpointDir(checkpointDir, checkpoint, container.Name, container.ID, container.CheckpointDir(), false)
		if err != nil {
			return err
		}
	}

	// 构建libcontainerd创建容器需要的参数
	createOptions, err := daemon.getLibcontainerdCreateOptions(container)
	if err != nil {
		return err
	}

	// 调用 libcontainerd.Client.Create() 通知containerd创建容器
	err = daemon.containerd.Create(context.Background(), container.ID, spec, createOptions)
	if err != nil {
		return translateContainerdStartErr(container.Path, container.SetExitCode, err)
	}

	// 调用 libcontainerd.Client.Start() 启动容器
	// TODO(mlaventure): we need to specify checkpoint options here
	pid, err := daemon.containerd.Start(context.Background(), container.ID, checkpointDir,
		container.StreamConfig.Stdin() != nil || container.Config.Tty,
		container.InitializeStdio)
	if err != nil {
		if err := daemon.containerd.Delete(context.Background(), container.ID); err != nil {
			logrus.WithError(err).WithField("container", container.ID).
				Error("failed to delete failed start container")
		}
		return translateContainerdStartErr(container.Path, container.SetExitCode, err)
	}

	// 设置container状态为Run, 记录pid（这里记录StartTime）
	container.SetRunning(pid, true)
	container.HasBeenManuallyStopped = false
	container.HasBeenStartedBefore = true
	daemon.setStateCounter(container)

	// 启动容器health check
	daemon.initHealthMonitor(container)

	if err := container.CheckpointTo(daemon.containersReplica); err != nil {
		logrus.WithError(err).WithField("container", container.ID).
			Errorf("failed to store container")
	}

	daemon.LogContainerEvent(container, "start")
	containerActions.WithValues("start").UpdateSince(start)

	return nil
}

// Cleanup releases any network resources allocated to the container along with any rules
// around how containers are linked together.  It also unmounts the container's root filesystem.
func (daemon *Daemon) Cleanup(container *container.Container) {
	// 清理网络相关
	daemon.releaseNetwork(container)

	// 取消ipc的挂载(shm目录的挂载)
	if err := container.UnmountIpcMount(detachMounted); err != nil {
		logrus.Warnf("%s cleanup: failed to unmount IPC: %s", container.ID, err)
	}

	// container rwlayer的umount
	if err := daemon.conditionalUnmountOnCleanup(container); err != nil {
		// FIXME: remove once reference counting for graphdrivers has been refactored
		// Ensure that all the mounts are gone
		if mountid, err := daemon.imageService.GetLayerMountID(container.ID, container.OS); err == nil {
			daemon.cleanupMountsByID(mountid)
		}
	}

	// 取消容器 secrets目录挂载
	if err := container.UnmountSecrets(); err != nil {
		logrus.Warnf("%s cleanup: failed to unmount secrets: %s", container.ID, err)
	}

	// 取消容器对应 元信息目录下的所有挂载
	if err := mount.RecursiveUnmount(container.Root); err != nil {
		logrus.WithError(err).WithField("container", container.ID).Warn("Error while cleaning up container resource mounts.")
	}

	// 删除所有 exec 的执行
	for _, eConfig := range container.ExecCommands.Commands() {
		daemon.unregisterExecCommand(container, eConfig)
	}

	// 取消所有 volume 挂载
	if container.BaseFS != nil && container.BaseFS.Path() != "" {
		if err := container.UnmountVolumes(daemon.LogVolumeEvent); err != nil {
			logrus.Warnf("%s cleanup: Failed to umount volumes: %v", container.ID, err)
		}
	}

	container.CancelAttachContext()

	// 调用 containerd.Delete() 停止删除containerd中对应container的记录
	if err := daemon.containerd.Delete(context.Background(), container.ID); err != nil {
		logrus.Errorf("%s cleanup: failed to delete container from containerd: %v", container.ID, err)
	}
}
