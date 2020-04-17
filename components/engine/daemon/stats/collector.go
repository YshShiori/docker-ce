package stats // import "github.com/docker/docker/daemon/stats"

import (
	"bufio"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/container"
	"github.com/docker/docker/pkg/pubsub"
	"github.com/sirupsen/logrus"
)

// Collector manages and provides container resource stats
type Collector struct {
	m          sync.Mutex
	supervisor supervisor
	interval   time.Duration
	publishers map[*container.Container]*pubsub.Publisher
	bufReader  *bufio.Reader

	// The following fields are not set on Windows currently.
	clockTicksPerSecond uint64
}

// NewCollector creates a stats collector that will poll the supervisor with the specified interval
func NewCollector(supervisor supervisor, interval time.Duration) *Collector {
	s := &Collector{
		interval:   interval,
		supervisor: supervisor,
		publishers: make(map[*container.Container]*pubsub.Publisher),
		bufReader:  bufio.NewReaderSize(nil, 128),
	}

	platformNewStatsCollector(s)

	return s
}

type supervisor interface {
	// GetContainerStats collects all the stats related to a container
	GetContainerStats(container *container.Container) (*types.StatsJSON, error)
}

// Collect registers the container with the collector and adds it to
// the event loop for collection on the specified interval returning
// a channel for the subscriber to receive on.
func (s *Collector) Collect(c *container.Container) chan interface{} {
	s.m.Lock()
	defer s.m.Unlock()
	// 容器第一次被订阅, 创建对应的publisher对象
	publisher, exists := s.publishers[c]
	if !exists {
		publisher = pubsub.NewPublisher(100*time.Millisecond, 1024)
		s.publishers[c] = publisher
	}
	// 订阅publisher的所有信息, 通过返回值的channel推送
	return publisher.Subscribe()
}

// StopCollection closes the channels for all subscribers and removes
// the container from metrics collection.
func (s *Collector) StopCollection(c *container.Container) {
	s.m.Lock()
	if publisher, exists := s.publishers[c]; exists {
		publisher.Close()
		delete(s.publishers, c)
	}
	s.m.Unlock()
}

// Unsubscribe removes a specific subscriber from receiving updates for a container's stats.
func (s *Collector) Unsubscribe(c *container.Container, ch chan interface{}) {
	s.m.Lock()
	publisher := s.publishers[c]
	if publisher != nil {
		// 从publisher中移除对应channel
		publisher.Evict(ch)
		// 如果没有人订阅了, 移除对应publisher
		if publisher.Len() == 0 {
			delete(s.publishers, c)
		}
	}
	s.m.Unlock()
}

// Run starts the collectors and will indefinitely collect stats from the supervisor
func (s *Collector) Run() {
	type publishersPair struct {
		container *container.Container
		publisher *pubsub.Publisher
	}
	// we cannot determine the capacity here.
	// it will grow enough in first iteration
	var pairs []publishersPair

	for {
		// 间隔时间
		// Put sleep at the start so that it will always be hit,
		// preventing a tight loop if no stats are collected.
		time.Sleep(s.interval)

		// pair对清零
		// it does not make sense in the first iteration,
		// but saves allocations in further iterations
		pairs = pairs[:0]

		// 读取<publishers>中所有的container与publisher对应
		s.m.Lock()
		for container, publisher := range s.publishers {
			// copy pointers here to release the lock ASAP
			pairs = append(pairs, publishersPair{container, publisher})
		}
		s.m.Unlock()
		if len(pairs) == 0 {
			continue
		}

		// 读取host online cpu数量
		onlineCPUs, err := s.getNumberOnlineCPUs()
		if err != nil {
			logrus.Errorf("collecting system online cpu count: %v", err)
			continue
		}

		// 对于每个pair(其实就是每个container)
		for _, pair := range pairs {
			// 通过 supervisor.GetContainerStats() 得到container对应stat信息
			stats, err := s.supervisor.GetContainerStats(pair.container)

			switch err.(type) {
			case nil:
				// 读取systemcpu, 记录到stats
				// Sample system CPU usage close to container usage to avoid
				// noise in metric calculations.
				systemUsage, err := s.getSystemCPUUsage()
				if err != nil {
					logrus.WithError(err).WithField("container_id", pair.container.ID).Errorf("collecting system cpu usage")
					continue
				}

				// FIXME: move to containerd on Linux (not Windows)
				stats.CPUStats.SystemUsage = systemUsage
				stats.CPUStats.OnlineCPUs = onlineCPUs

				// 推送stats
				pair.publisher.Publish(*stats)

			case notRunningErr, notFoundErr:
				// 对应container不存在, 推送 零值的stat信息
				// publish empty stats containing only name and ID if not running or not found
				pair.publisher.Publish(types.StatsJSON{
					Name: pair.container.Name,
					ID:   pair.container.ID,
				})

			default:
				logrus.Errorf("collecting stats for %s: %v", pair.container.ID, err)
			}
		}
	}
}

type notRunningErr interface {
	error
	Conflict()
}

type notFoundErr interface {
	error
	NotFound()
}
