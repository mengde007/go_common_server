package common

import (
	"fmt"
	"logger"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

func DebugEnd(cpuProfile bool) {
	if cpuProfile {
		pprof.StopCPUProfile()
	}
}

func DebugInit(gcTime uint8, debugDns string, proccessname string, cpuProfile bool) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	os.Setenv("GOTRACEBACK", "crash")

	if cpuProfile {
		fullname := fmt.Sprintf("CpuProfile%s%d", proccessname, time.Now().Unix())
		f, err := os.Create(fullname)
		if err != nil {
			logger.Fatal("os.Create err", err)
		}
		pprof.StartCPUProfile(f)
	}

	if gcTime > 60 {
		gcTime = 15
	}

	go func() {
		var m runtime.MemStats
		for {
			//HeapSys：程序向应用程序申请的内存
			//HeapAlloc：堆上目前分配的内存
			//HeapIdle：堆上目前没有使用的内存
			//HeapReleased：回收到操作系统的内存
			runtime.GC()
			debug.FreeOSMemory()
			runtime.ReadMemStats(&m)
			// logger.Info("Gc : HeapSys =%v, HeapAlloc =%v, HeapIdle=%v, HeapReleased=%v", m.HeapSys, m.HeapAlloc,
			// 	m.HeapIdle, m.HeapReleased)
			time.Sleep(time.Second * time.Duration(gcTime))
		}
	}()

	go func() {
		logger.Info("Debug Http Service : %v", debugDns)
		logger.Info("Listen:", http.ListenAndServe(debugDns, nil))
	}()
}
