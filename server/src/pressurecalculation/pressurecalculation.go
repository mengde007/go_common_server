package pressurecalculation

import (
	"logger"
	"sort"
	"sync"
	"time"
	"timer"
)

type Func struct {
	name      string
	m         map[int64]int64
	timeBegin int64
	timeEnd   int64
	iAutoId   int64
	l         sync.Mutex
	times     int64
	timeMin   int64
	timeMax   int64
	ints      []int
	tm        *timer.Timer
}

func (self *Func) init() {
	self.tm = timer.NewTimer(time.Second * 10)
	self.tm.Start(func() {
		self.l.Lock()
		defer self.l.Unlock()

		if self.times == int64(0) {
			return
		}

		num := len(self.ints) * 9 / 10
		if num == 0 {
			return
		}

		sort.Ints(self.ints)

		total := int64(0)
		for i := 0; i < num; i++ {
			total += int64(self.ints[i])
		}

		logger.Error("%s pressurecalculation : totaltimes, failed, avgnumber, mintime, maxtime, 90%%time:",
			self.name,
			self.times,
			len(self.m),
			self.times*int64(time.Second)/(self.timeEnd-self.timeBegin),
			self.timeMin/int64(time.Millisecond),
			self.timeMax/int64(time.Millisecond),
			total/int64(num)/int64(time.Millisecond),
		)
	})
}

func (self *Func) enter() int64 {
	self.l.Lock()
	self.iAutoId++
	id := self.iAutoId
	self.m[id] = time.Now().UnixNano()
	self.l.Unlock()

	return id
}

func (self *Func) leave(id int64) {
	timeNow := time.Now().UnixNano()
	self.l.Lock()
	if timeBegin, ok := self.m[id]; ok {
		diff := timeNow - timeBegin

		self.times++

		if diff > self.timeMax {
			self.timeMax = diff
		}

		if self.timeMin == int64(0) || diff < self.timeMin {
			self.timeMin = diff
		}

		self.ints = append(self.ints, int(diff))
		delete(self.m, id)
	}
	self.timeEnd = timeNow
	self.l.Unlock()
}

var mapFuncs map[string]*Func = make(map[string]*Func)
var l sync.RWMutex
var open bool = false

func OnEnterFunc(fun string) int64 {
	if !open {
		return 0
	}

	l.RLock()
	if m, ok := mapFuncs[fun]; ok {
		l.RUnlock()
		return m.enter()
	}
	l.RUnlock()

	l.Lock()
	if m, ok := mapFuncs[fun]; ok {
		l.Unlock()
		return m.enter()
	}

	m := &Func{
		name:      fun,
		m:         make(map[int64]int64),
		timeBegin: time.Now().UnixNano(),
		iAutoId:   int64(0),
		times:     int64(0),
		timeMin:   int64(0),
		timeMax:   int64(0),
		ints:      make([]int, 0),
	}
	m.init()
	mapFuncs[fun] = m
	l.Unlock()

	return m.enter()
}

func OnLeaveFunc(fun string, id int64) {
	if !open {
		return
	}

	l.RLock()
	if m, ok := mapFuncs[fun]; ok {
		m.leave(id)
	}
	l.RUnlock()
}

func Clean() {
	l.Lock()
	defer l.Unlock()

	for _, fun := range mapFuncs {
		fun.l.Lock()
		fun.tm.Stop()
		fun.l.Unlock()
	}

	mapFuncs = make(map[string]*Func)
}

func Open(bOpen bool) {
	l.Lock()
	open = bOpen
	l.Unlock()
}
