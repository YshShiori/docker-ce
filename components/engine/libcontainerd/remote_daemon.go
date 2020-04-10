// +build !windows

package libcontainerd // import "github.com/docker/docker/libcontainerd"

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/server"
	"github.com/docker/docker/pkg/system"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	maxConnectionRetryCount = 3
	healthCheckTimeout      = 3 * time.Second
	shutdownTimeout         = 15 * time.Second
	configFile              = "containerd.toml"
	binaryName              = "docker-containerd"
	pidFile                 = "docker-containerd.pid"
)

type pluginConfigs struct {
	Plugins map[string]interface{} `toml:"plugins"`
}

type remote struct {
	sync.RWMutex
	server.Config

	daemonPid int
	logger    *logrus.Entry

	daemonWaitCh    chan struct{}
	clients         []*client
	shutdownContext context.Context
	shutdownCancel  context.CancelFunc
	shutdown        bool

	// Options
	startDaemon bool
	rootDir     string
	stateDir    string
	snapshotter string
	pluginConfs pluginConfigs
}

// New creates a fresh instance of libcontainerd remote.
func New(rootDir, stateDir string, options ...RemoteOption) (rem Remote, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "Failed to connect to containerd")
		}
	}()

	// 构造remote对象
	r := &remote{
		rootDir:  rootDir,
		stateDir: stateDir,
		Config: server.Config{
			Root:  filepath.Join(rootDir, "daemon"),
			State: filepath.Join(stateDir, "daemon"),
		},
		pluginConfs: pluginConfigs{make(map[string]interface{})},
		daemonPid:   -1,
		logger:      logrus.WithField("module", "libcontainerd"),
	}
	r.shutdownContext, r.shutdownCancel = context.WithCancel(context.Background())

	// opts处理
	rem = r
	for _, option := range options {
		if err = option.Apply(r); err != nil {
			return
		}
	}

	// 设置一些默认参数
	r.setDefaults()

	// 创建stateDir目录
	if err = system.MkdirAll(stateDir, 0700, ""); err != nil {
		return
	}

	// 如果指定<startDaemon>, 启动containerd
	if r.startDaemon {
		os.Remove(r.GRPC.Address)
		if err = r.startContainerd(); err != nil {
			return
		}
		defer func() {
			if err != nil {
				r.Cleanup()
			}
		}()
	}

	// 创建containerd.Client, 调用Client.Version()测试一下连接
	// This connection is just used to monitor the connection
	client, err := containerd.New(r.GRPC.Address)
	if err != nil {
		return
	}
	if _, err := client.Version(context.Background()); err != nil {
		system.KillProcess(r.daemonPid)
		return nil, errors.Wrapf(err, "unable to get containerd version")
	}

	// 通过这个client, 周期性监控containerd是否健康
	go r.monitorConnection(client)

	return r, nil
}

func (r *remote) NewClient(ns string, b Backend) (Client, error) {
	// 创建client实例
	c := &client{
		stateDir:   r.stateDir,
		logger:     r.logger.WithField("namespace", ns),
		namespace:  ns,
		backend:    b,
		containers: make(map[string]*container),
	}

	// 创建container.Client, 赋值到[client].remote
	rclient, err := containerd.New(r.GRPC.Address, containerd.WithDefaultNamespace(ns))
	if err != nil {
		return nil, err
	}
	c.remote = rclient

	// 通过 container.Client.EventService().Subscribe() 订阅namespace下的所有event
	go c.processEventStream(r.shutdownContext)

	// 记录到<clients>
	r.Lock()
	r.clients = append(r.clients, c)
	r.Unlock()
	return c, nil
}

func (r *remote) Cleanup() {
	if r.daemonPid != -1 {
		r.shutdownCancel()
		r.stopDaemon()
	}

	// cleanup some files
	os.Remove(filepath.Join(r.stateDir, pidFile))

	r.platformCleanup()
}

func (r *remote) getContainerdPid() (int, error) {
	pidFile := filepath.Join(r.stateDir, pidFile)
	f, err := os.OpenFile(pidFile, os.O_RDWR, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, nil
		}
		return -1, err
	}
	defer f.Close()

	b := make([]byte, 8)
	n, err := f.Read(b)
	if err != nil && err != io.EOF {
		return -1, err
	}

	if n > 0 {
		pid, err := strconv.ParseUint(string(b[:n]), 10, 64)
		if err != nil {
			return -1, err
		}
		if system.IsProcessAlive(int(pid)) {
			return int(pid), nil
		}
	}

	return -1, nil
}

func (r *remote) getContainerdConfig() (string, error) {
	path := filepath.Join(r.stateDir, configFile)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open containerd config file at %s", path)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err = enc.Encode(r.Config); err != nil {
		return "", errors.Wrapf(err, "failed to encode general config")
	}
	if err = enc.Encode(r.pluginConfs); err != nil {
		return "", errors.Wrapf(err, "failed to encode plugin configs")
	}

	return path, nil
}

func (r *remote) startContainerd() error {
	// 读取已经运行着的container pid
	pid, err := r.getContainerdPid()
	if err != nil {
		return err
	}

	if pid != -1 {
		r.daemonPid = pid
		logrus.WithField("pid", pid).
			Infof("libcontainerd: %s is still running", binaryName)
		return nil
	}

	configFile, err := r.getContainerdConfig()
	if err != nil {
		return err
	}

	// 启动containerd!
	args := []string{"--config", configFile}
	cmd := exec.Command(binaryName, args...)
	// redirect containerd logs to docker logs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = containerdSysProcAttr()
	// clear the NOTIFY_SOCKET from the env when starting containerd
	cmd.Env = nil
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "NOTIFY_SOCKET") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	// 启动goroutine, 阻塞等待containerd进程退出, 并通过<daemonWaitCh>通知
	r.daemonWaitCh = make(chan struct{})
	go func() {
		// Reap our child when needed
		if err := cmd.Wait(); err != nil {
			r.logger.WithError(err).Errorf("containerd did not exit successfully")
		}
		close(r.daemonWaitCh)
	}()

	// 记录pid至<daemonPid>
	r.daemonPid = cmd.Process.Pid

	// 将pid写入到"docker-containerd.pid"文件
	err = ioutil.WriteFile(filepath.Join(r.stateDir, pidFile), []byte(fmt.Sprintf("%d", r.daemonPid)), 0660)
	if err != nil {
		system.KillProcess(r.daemonPid)
		return errors.Wrap(err, "libcontainerd: failed to save daemon pid to disk")
	}

	logrus.WithField("pid", r.daemonPid).
		Infof("libcontainerd: started new %s process", binaryName)

	return nil
}

func (r *remote) monitorConnection(monitor *containerd.Client) {
	var transientFailureCount = 0

	for {
		select {
		case <-r.shutdownContext.Done():
			// 收到关闭remote请求, 关闭client
			r.logger.Info("stopping healthcheck following graceful shutdown")
			monitor.Close()
			return
		case <-time.After(500 * time.Millisecond):
		}

		// 通过 client.IsServing() 检查containerd是否健康
		ctx, cancel := context.WithTimeout(r.shutdownContext, healthCheckTimeout)
		_, err := monitor.IsServing(ctx)
		cancel()
		if err == nil {
			transientFailureCount = 0
			continue
		}

		select {
		case <-r.shutdownContext.Done():
			r.logger.Info("stopping healthcheck following graceful shutdown")
			monitor.Close()
			return
		default:
		}

		r.logger.WithError(err).WithField("binary", binaryName).Debug("daemon is not responding")

		if r.daemonPid == -1 {
			continue
		}

		// 如果失败次数超过阈值, kill containerd进程
		transientFailureCount++
		if transientFailureCount < maxConnectionRetryCount || system.IsProcessAlive(r.daemonPid) {
			continue
		}

		transientFailureCount = 0
		if system.IsProcessAlive(r.daemonPid) {
			r.logger.WithField("pid", r.daemonPid).Info("killing and restarting containerd")
			// Try to get a stack trace
			syscall.Kill(r.daemonPid, syscall.SIGUSR1)
			<-time.After(100 * time.Millisecond)
			system.KillProcess(r.daemonPid)
		}
		<-r.daemonWaitCh

		// 重新启动containerd进程
		monitor.Close()
		os.Remove(r.GRPC.Address)
		if err := r.startContainerd(); err != nil {
			r.logger.WithError(err).Error("failed restarting containerd")
			continue
		}

		// 重新构建client
		newMonitor, err := containerd.New(r.GRPC.Address)
		if err != nil {
			r.logger.WithError(err).Error("failed connect to containerd")
			continue
		}

		monitor = newMonitor
		var wg sync.WaitGroup

		// 所有client重新赋值新的连接
		// WARN: 这里当前版本应该是有bug, container等结构中client可能不会更新
		for _, c := range r.clients {
			wg.Add(1)

			go func(c *client) {
				defer wg.Done()
				c.logger.WithField("namespace", c.namespace).Debug("creating new containerd remote client")
				c.remote.Close()

				// 创建新的containerd.Client
				remote, err := containerd.New(r.GRPC.Address, containerd.WithDefaultNamespace(c.namespace))
				if err != nil {
					r.logger.WithError(err).Error("failed to connect to containerd")
					// TODO: Better way to handle this?
					// This *shouldn't* happen, but this could wind up where the daemon
					// is not able to communicate with an eventually up containerd
					return
				}

				// 赋值到client.remote
				c.setRemote(remote)
			}(c)

			wg.Wait()
		}
	}
}
