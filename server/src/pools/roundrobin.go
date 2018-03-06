//	管理可服用资源， 比如连接池。
package pools

import (
	"fmt"
	"sync"
	"time"
)

type Factory func() (Resource, error)

type Resource interface {
	Close()
	IsClosed() bool
}

type RoundRobin struct {
	mu          sync.Mutex
	available   *sync.Cond
	resources   chan fifoWrapper
	size        int64
	factory     Factory
	idleTimeout time.Duration

	waitCount int64
	waitTime  time.Duration
}

type fifoWrapper struct {
	resource Resource
	timeUsed time.Time
}

func NewRoundRobin(capacity int, idleTimeout time.Duration) *RoundRobin {
	r := &RoundRobin{
		resources:   make(chan fifoWrapper, capacity),
		size:        0,
		idleTimeout: idleTimeout,
	}
	r.available = sync.NewCond(&r.mu)
	return r
}

func (self *RoundRobin) Open(factory Factory) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.factory = factory
}

func (self *RoundRobin) Close() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for self.size > 0 {
		select {
		case fw := <-self.resources:
			go fw.resource.Close()
			self.size--
		default:
			self.available.Wait()
		}
	}
	self.factory = nil
}

func (self *RoundRobin) IsClosed() bool {
	return self.factory == nil
}

func (self *RoundRobin) Get() (resource Resource, err error) {
	return self.get(true)
}

func (self *RoundRobin) TryGet() (resource Resource, err error) {
	return self.get(false)
}

func (self *RoundRobin) get(wait bool) (resource Resource, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	for {
		select {
		case fw := <-self.resources:

			if self.idleTimeout > 0 && fw.timeUsed.Add(self.idleTimeout).Sub(time.Now()) < 0 {

				go fw.resource.Close()
				self.size--

				self.available.Signal()
				continue
			}
			return fw.resource, nil
		default:

			if self.size >= int64(cap(self.resources)) {
				if wait {
					start := time.Now()
					self.available.Wait()
					self.recordWait(start)
					continue
				}
				return nil, nil
			}

			if resource, err = self.waitForCreate(); err != nil {

				self.available.Signal()
				return nil, err
			}

			self.size++
			return resource, err
		}
	}
	panic("unreachable")
}

func (self *RoundRobin) recordWait(start time.Time) {
	self.waitCount++
	self.waitTime += time.Now().Sub(start)
}

func (self *RoundRobin) waitForCreate() (resource Resource, err error) {

	self.size++
	self.mu.Unlock()
	defer func() {
		self.mu.Lock()
		self.size--
	}()
	return self.factory()
}

func (self *RoundRobin) Put(resource Resource) {
	self.mu.Lock()
	defer self.available.Signal()
	defer self.mu.Unlock()

	if self.size > int64(cap(self.resources)) {
		go resource.Close()
		self.size--
	} else if resource.IsClosed() {
		self.size--
	} else {
		if len(self.resources) == cap(self.resources) {
			panic("unexpected")
		}
		self.resources <- fifoWrapper{resource, time.Now()}
	}
}

func (self *RoundRobin) SetCapacity(capacity int) {
	self.mu.Lock()
	defer self.available.Broadcast()
	defer self.mu.Unlock()

	nr := make(chan fifoWrapper, capacity)

	for {
		select {
		case fw := <-self.resources:
			if len(nr) < cap(nr) {
				nr <- fw
			} else {
				go fw.resource.Close()
				self.size--
			}
			continue
		default:
		}
		break
	}
	self.resources = nr
}

func (self *RoundRobin) SetIdleTimeout(idleTimeout time.Duration) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.idleTimeout = idleTimeout
}

func (self *RoundRobin) StatsJSON() string {
	s, c, a, wc, wt, it := self.Stats()
	return fmt.Sprintf("{\"Size\": %v, \"Capacity\": %v, \"Available\": %v, \"WaitCount\": %v, \"WaitTime\": %v, \"IdleTimeout\": %v}", s, c, a, wc, int64(wt), int64(it))
}

func (self *RoundRobin) Stats() (size, capacity, available, waitCount int64, waitTime, idleTimeout time.Duration) {
	self.mu.Lock()
	defer self.mu.Unlock()
	return self.size, int64(cap(self.resources)), int64(len(self.resources)), self.waitCount, self.waitTime, self.idleTimeout
}
