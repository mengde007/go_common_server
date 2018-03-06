package main

import (
	"centerclient"
	"common"
	"connector"
	"flag"
	"logger"
	"net"
	_ "net/http/pprof"
	"os"
	"rpcplus"
	"sync"
	"syscall"
)

var (
	//	laddr    = flag.String("l", "192.168.8.103:8820", "The address to bind to.")
	//	dbg_addr = flag.String("d", "127.0.0.1:8821", "The address to bind to.(for debug)")
	csvDir = flag.String("c", "config", "config dir")
)

func main() {
	logger.Info("cnserver start")

	flag.Parse()

	if err := common.ReadCnsServerConfig(*csvDir, &connector.Cfg); err != nil {
		logger.Fatal("load cns config error", *csvDir, err)
		return
	}

	common.DebugInit(connector.Cfg.GcTime, connector.Cfg.DebugHost, "gameserver", connector.Cfg.CpuProfile)

	//dbg_sock, err := net.Listen("tcp", *dbg_addr)
	//if err != nil {
	//	logger.Fatal("net.Listen: %s", err.Error())
	//}
	//go http.Serve(dbg_sock, nil)

	cnServer := connector.NewCNServer(&connector.Cfg)

	handler := func(s os.Signal, arg interface{}) {
		common.DebugEnd(connector.Cfg.CpuProfile)
		logger.Error("cnserver handle signal: %d", s)
		cnServer.Quit()
		logger.Error("cnserver will close")
	}

	handlerArray := []os.Signal{syscall.SIGINT,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGKILL}

	logger.Error("WatchSystemSignal!!!!!!!!!!!")
	go common.WatchSystemSignal(&handlerArray, handler)

	wg := &sync.WaitGroup{}
	cnServer.StartClientService(&connector.Cfg, wg)

	//已经监听过了
	var listener net.Listener = nil

	fCallback := func(conn *rpcplus.Client) {
		logger.Info("on center connected !")

		if listener != nil {
			connector.StartCenterService(cnServer, listener, &connector.Cfg)
		} else {
			iTimes := 0
			for {
				iTimes++

				//尝试次数
				if iTimes > 5 {
					logger.Fatal("listen for center failed !")
					break
				}

				csock, err := net.Listen("tcp", connector.Cfg.CnsForCenter)
				if err != nil {
					logger.Error("net.Listen: %s", err.Error())
					continue
				}

				listener = csock

				connector.StartCenterService(cnServer, csock, &connector.Cfg)
				break
			}
		}
	}
	//连接center成功后的回调
	centerclient.SetConnectedCallback(fCallback)
	//第一次手动调用
	go fCallback(nil)

	//这个放到最后，因为要wait所有的客户端下线保存数据
	logger.Error("wait client Quit!!!!!!!!!!!")
	wg.Wait()
	logger.Error("all client Quit!!!!!!!!!!!")
	cnServer.EndService()
	logger.Error("cnserver end")
}
