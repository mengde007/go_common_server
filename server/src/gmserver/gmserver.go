package gmserver

import (
	"centerclient"
	"common"
	"connector"
	// "dbclient"
	"errors"
	"fmt"
	"io/ioutil"
	"lockclient"
	"logger"
	"net"
	"net/http"
	"proto"
	// "rpc"
	"rpcplus"
	"rpcplusclientpool"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

var serverType uint32

type GmService struct {
	pPoll      *rpcplusclientpool.ClientPool
	pCachePool *common.CachePool
}

var pGmService *GmService

func CreateGmServer() {
	//center
	aServerHost := common.ReadServerClientConfig("centerservergm")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll := rpcplusclientpool.CreateClientPool(aServerHost, "gmserver -> center")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	var cfg common.GmServerCfg
	if err := common.ReadGmConfig(&cfg); err != nil {
		logger.Error("ReadGmConfig failed", err)
		return
	}
	// if cfg.ServerType != ServerType_QQ && cfg.ServerType != ServerType_WX {
	// 	logger.Fatal("wrong server type")
	// 	return
	// }
	serverType = cfg.ServerType

	pGmService = &GmService{
		pPoll:      pPoll,
		pCachePool: common.NewCachePool(cfg.Maincache),
	}

	//配置表
	connector.LoadConfigFiles(common.GetDesignerDir())
	common.LoadGlobalConfig()
	common.LoadDaerGlobalConfig()

	//通知服务
	go pGmService.startNoticeService()

	wg := sync.WaitGroup{}
	wg.Add(2)

	//监听内网
	go pGmService.initTcp(&cfg, &wg)
	//监听网页
	go pGmService.initHttp(&cfg, &wg)

	wg.Wait()
}

func (self *GmService) initTcp(cfg *common.GmServerCfg, wg *sync.WaitGroup) error {
	defer wg.Done()

	//监听
	listener, err := net.Listen("tcp", cfg.InnerHost)
	if err != nil {
		logger.Error("Listening to: %s %s", cfg.InnerHost, " failed !!")
		return err
	}
	defer listener.Close()

	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pGmService)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gmserver StartServices %s", err.Error())
			break
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("gmserver Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()

			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}

	return nil
}

//回调函数创建
type handleCALLBACK func(w http.ResponseWriter, r *http.Request)

func createHandleFunc(f handleCALLBACK) handleCALLBACK {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				errmsg := fmt.Sprintf("handle http failed :%s", r)
				writeString(w, "serious err occurred :", errmsg)
				return
			}
		}()

		f(w, r)
	}
}

const (
	REQHEAD = "data_packet="
)

func (self *GmService) handle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("ReadAll body err", err)
		return
	}

	data := string(b)
	// if len(data) < len(REQHEAD) || data[:len(REQHEAD)] != REQHEAD {
	// 	logger.Error("wrong request string")
	// 	return
	// }

	// data = data[len(REQHEAD):]
	// if data == "" {
	// 	logger.Error("no data_packet")
	// 	return
	// }

	logger.Info("request data:", data)

	head, err := ParseCmdHeadNew(data)
	if err != nil || head == nil {
		logger.Error("ParseCmdHead failed", err)
		return
	}

	cmdId := head.Commid
	switch cmdId {
	case 5002: //角色人物信息修改
		self.handle_modify_role_info(w, head, data)
	case 5005: //邮件
		self.handle_send_mail(w, head, data)
	case 5008: //创建轮播
		self.handle_notice_add(w, head, data)
	case 5009: //删除轮播
		self.handle_delete_notice(w, head, data)
	case 5010: //统计信息
		self.handle_accounting_info(w, head, data)

	default:
		logger.Error("wrong cmdId:", cmdId)
	}
}

func (self *GmService) initHttp(cfg *common.GmServerCfg, wg *sync.WaitGroup) error {
	defer wg.Done()

	http.HandleFunc("/", createHandleFunc(self.handle))
	// http.HandleFunc("/account/forceunlock", createHandleFunc(self.handle_account_forceunlock))
	// http.HandleFunc("/account/name", createHandleFunc(self.handle_old_account_name))
	// http.HandleFunc("/account/unbind", createHandleFunc(self.handle_account_unbind))

	//对外
	if err := http.ListenAndServe(cfg.Host, nil); err != nil {
		return err
	}

	return nil
}

//更新openid和名字
func (self *GmService) UpdateOpenId2Name(req *proto.GmUpdateOpenId2Name, rst *proto.GmUpdateOpenId2Name) error {
	keyOld := req.OpenId + ":" + strconv.Itoa(int(req.Platform)) + ":" + req.NameLast
	keyNew := req.OpenId + ":" + strconv.Itoa(int(req.Platform)) + ":" + req.Name

	cache := self.pCachePool.Get()
	defer cache.Recycle()

	if keyOld != keyNew {
		cache.Do("SREM", common.TableOpenId2Name, keyOld)
	}
	cache.Do("SADD", common.TableOpenId2Name, keyNew)

	return nil
}

func (self *GmService) lockPlayer(uid string) (uint64, error) {
	lid := common.GenLockMessage(common.Special_Server_Id, proto.MethodPlayerGmOpera, 0)
	ok, value, err := lockclient.TryLock(common.LockName_Player, uid, lid, common.LockTime_GM, nil)
	if err != nil {
		return 0, err
	}
	if ok {
		return lid, nil
	}

	_, tid, _, _, _ := common.ParseLockMessage(value)
	if tid != proto.MethodPlayerLogin {
		return 0, errors.New("can not lock player")
	}

	req := &proto.LoginKickPlayer{
		Id: uid,
	}
	rst := &proto.LoginKickPlayerResult{}
	if err := centerclient.Call("Center.KickCnsPlayer", req, rst); err == nil && rst.Success {
	} else {
		return 0, errors.New("kick failed")
	}

	//等100毫秒
	time.Sleep(time.Millisecond * 100)

	//再试一次
	ok, value, err = lockclient.TryLock(common.LockName_Player, uid, lid, common.LockTime_GM, nil)
	if err != nil {
		return 0, err
	}

	if ok {
		return lid, nil
	}

	return 0, errors.New("lock failed")
}

func (self *GmService) unlockPlayer(uid string, lid uint64) error {
	_, err := lockclient.TryUnlock(common.LockName_Player, uid, lid)

	return err
}

//强制解锁
func (self *GmService) forceUnlockPlayer(uid string) error {
	ok, err := lockclient.ForceUnLock(common.LockName_Player, uid)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("opera failed")
	}

	return nil
}

//gm取在线玩家数量
func (self *GmService) GmGetOnlineNum(req *proto.GmGetOnlineNum, rst *proto.GmGetOnlineNumResult) error {
	logger.Info("GmGetOnlineNum:")
	err, conn := self.pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	return conn.Call("CenterGmServices.GmGetOnlineNum", req, rst)
}

func (self *GmService) GmCheckActivityConfig(req *proto.GmCheckActivityConfig, rst *proto.GmCheckActivityConfigResult) error {
	logger.Info("come into GmCheckActivityConfig")
	err, conn := self.pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	return conn.Call("CenterGmServices.ReloadCfg", req, rst)
}
