package pools

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Numbered struct {
	mu        sync.Mutex
	empty     *sync.Cond
	resources map[int64]*numberedWrapper
}

type numberedWrapper struct {
	val         interface{}
	inUse       bool
	timeCreated time.Time
}

func NewNumbered() *Numbered {
	n := &Numbered{resources: make(map[int64]*numberedWrapper)}
	n.empty = sync.NewCond(&n.mu)
	return n
}

func (self *Numbered) Register(id int64, val interface{}) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if _, ok := self.resources[id]; ok {
		return errors.New("already present")
	}
	self.resources[id] = &numberedWrapper{val, false, time.Now()}
	return nil
}

func (self *Numbered) Unregister(id int64) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.resources, id)
	if len(self.resources) == 0 {
		self.empty.Broadcast()
	}
}

func (self *Numbered) Get(id int64) (val interface{}, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	nw, ok := self.resources[id]
	if !ok {
		return nil, errors.New("not found")
	}
	if nw.inUse {
		return nil, errors.New("in use")
	}
	nw.inUse = true
	return nw.val, nil
}

func (self *Numbered) Put(id int64) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if nw, ok := self.resources[id]; ok {
		nw.inUse = false
	}
}

func (self *Numbered) GetTimedout(timeout time.Duration) (vals []interface{}) {
	self.mu.Lock()
	defer self.mu.Unlock()
	now := time.Now()
	for _, nw := range self.resources {
		if nw.inUse {
			continue
		}
		if nw.timeCreated.Add(timeout).Sub(now) <= 0 {
			nw.inUse = true
			vals = append(vals, nw.val)
		}
	}
	return vals
}

func (self *Numbered) WaitForEmpty() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for len(self.resources) != 0 {
		self.empty.Wait()
	}
}

func (self *Numbered) StatsJSON() string {
	s := self.Stats()
	return fmt.Sprintf("{\"Size\": %v}", s)
}

func (self *Numbered) Stats() (size int) {
	self.mu.Lock()
	defer self.mu.Unlock()
	return len(self.resources)
}
