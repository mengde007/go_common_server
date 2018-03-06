package common

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"reflect"
	"sync"
	//"time"
	"logger"
	"runtime/debug"
)

type methodType struct {
	method  reflect.Method
	ArgType reflect.Type
}

type cmdType struct {
	rtType reflect.Type
}

// 通讯包协议头，所有的头文件都要用这个
type cmdHead struct {
	MsgId  uint16 // 包ID，对应反序列化的json类型
	CmdLen uint16 // json数据段长度
}

type PpeConn struct {
	conn       net.Conn
	isShutDown bool
	shutDown   chan bool
	recvBuf    bytes.Buffer
	sendBuf    bytes.Buffer
	methods    map[string]*methodType
	msgs       map[uint16]*cmdType
	regitesr   interface{}
	lsend      sync.RWMutex
	send       chan bool
}

func CreatePpeConn(conn net.Conn, dispacth interface{}) *PpeConn {
	newPpeConn := &PpeConn{conn: conn,
		send:       make(chan bool, 100),
		shutDown:   make(chan bool),
		isShutDown: false}

	newPpeConn.register(dispacth)
	go newPpeConn.dispatchRoutine()
	go newPpeConn.sendRoutine()
	go newPpeConn.shutDownRoutine()
	return newPpeConn
}

func (self *PpeConn) shutDownRoutine() {
	<-self.shutDown
	self.conn.Close()
	logger.Error("shutDownRoutine")
}

func (self *PpeConn) SendUint16(data uint16) {
	if self.isShutDown {
		return
	}
	self.lsend.Lock()
	defer self.lsend.Unlock()
	binary.Write(&self.sendBuf, binary.BigEndian, data)
	self.send <- true
}

func (self *PpeConn) Send(data []byte) {
	if self.isShutDown {
		return
	}
	self.lsend.Lock()
	defer self.lsend.Unlock()

	self.sendBuf.Write(data)
	self.send <- true
}

func (self *PpeConn) ShutDown() {
	if self.isShutDown {
		return
	}
	self.isShutDown = true
	self.shutDown <- true
}

func (self *PpeConn) register(arg interface{}) {
	self.regitesr = arg
	self.methods = make(map[string]*methodType)
	self.msgs = make(map[uint16]*cmdType)

	// Install the methods
	rgType := reflect.TypeOf(arg)
	msgId := uint16(1)
	for c := 0; c < rgType.NumMethod(); c++ {

		m := rgType.Method(c)
		mt := m.Type
		numP := mt.NumIn()

		if numP != 3 {
			continue
		}

		argType := mt.In(1)
		connType := mt.In(2)

		if connType.Elem().Name() != "PpeConn" {
			continue
		}

		self.msgs[msgId] = &cmdType{rtType: argType.Elem()}
		self.methods[argType.Elem().Name()] = &methodType{method: m}
		msgId++
	}

	if len(self.msgs) == 0 {
	}
}

func (self *PpeConn) sendRoutine() {
	defer func() { self.ShutDown() }()
	for {
		<-self.send
		self.lsend.Lock()
		var tembuf bytes.Buffer
		tembuf.Write(self.sendBuf.Bytes())
		self.sendBuf.Reset()
		self.lsend.Unlock()

		_, err := tembuf.WriteTo(self.conn)

		if err != nil {
			break
		}
	}
}

func (self *PpeConn) dispatchRoutine() {
	defer func() {
		self.ShutDown()
		if r := recover(); r != nil {
			logger.Error("runtime error:", r)
			debug.PrintStack()
		}
	}()

	temp := make([]byte, 1024) //2kb的缓冲区
	for {
		nRead, err := self.conn.Read(temp)
		if err != nil {
			logger.Error("conn.Read error :%s", err)
			break
		}

		self.recvBuf.Write(temp[:nRead])
		ok := true
		for {
			if self.recvBuf.Len() < 4 {
				break
			}

			srcData := self.recvBuf.Bytes()
			msgId := []uint8{srcData[0], srcData[1]}
			dataLen := []uint8{srcData[2], srcData[3]}

			msgIdLE := binary.LittleEndian.Uint16(msgId)
			dataLenLE := binary.LittleEndian.Uint16(dataLen)

			if self.recvBuf.Len() < int(dataLenLE+4) {
				break
			}

			cmdType, ok := self.msgs[msgIdLE]

			if !ok {
				break
			}

			self.recvBuf.Next(4)
			rtType := cmdType.rtType
			ptr := reflect.New(rtType)

			err := json.Unmarshal(self.recvBuf.Bytes(), ptr.Interface())
			if err != nil {
				logger.Error("json.Unmarshal error :%s", err)
				ok = false
				break
			}
			self.methods[rtType.Name()].method.Func.Call([]reflect.Value{reflect.ValueOf(self.regitesr), ptr, reflect.ValueOf(self)})
			self.recvBuf.Next(int(dataLenLE))
		}

		if !ok {
			break
		}

	}
}
