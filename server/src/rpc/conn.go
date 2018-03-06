package rpc

import (
	"code.google.com/p/goprotobuf/proto"
	// "code.google.com/p/snappy-go/snappy"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"logger"
	"net"
	"reflect"
	"sync"
	"time"
	"timer"
)

const (
	ConnReadTimeOut  = 5e9
	ConnWriteTimeOut = 5e9
	ConnMaxBufSize   = 5120
)

type ProtoBufConn struct {
	c               net.Conn
	id              uint64
	send            chan *Request
	readbuf         []byte
	t               *timer.Timer
	exit            chan bool
	last_time       int64
	time_out        uint32
	lockForClose    sync.Mutex
	is_closed       bool
	lockChannelSize sync.Mutex
	channelSize     int32
	channelMaxSize  int32
	sync.Mutex
	connMgr *Server
}

func NewProtoBufConn(server *Server, c net.Conn, size int32, k uint32) (conn RpcConn) {
	pbc := &ProtoBufConn{
		c:              c,
		channelMaxSize: size,
		send:           make(chan *Request, size),
		exit:           make(chan bool, 1),
		readbuf:        make([]byte, ConnMaxBufSize),
		last_time:      time.Now().Unix(),
		time_out:       k,
		connMgr:        server,
		channelSize:    0,
	}

	if k > 0 {
		pbc.t = timer.NewTimer(time.Duration(k) * time.Second)
		pbc.t.Start(
			func() {
				pbc.OnCheck()
			},
		)
	}

	go pbc.mux()
	return pbc
}

func (conn *ProtoBufConn) OnCheck() {
	time_diff := uint32(time.Now().Unix() - conn.last_time)
	if time_diff > conn.time_out<<1 {
		//logger.Info("Conn %d TimeOut: %d", conn.GetId(), time_diff)
		//conn.connMgr.CloseConn(conn.GetId())
		conn.Close()
	}
}

func (conn *ProtoBufConn) mux() {
	for {
		select {
		case r := <-conn.send:

			conn.lockChannelSize.Lock()
			conn.channelSize--
			conn.lockChannelSize.Unlock()

			buf, err := proto.Marshal(r)
			if err != nil {
				logger.Error("ProtoBufConn Marshal Error %s", err.Error())
				continue
			}

			// dst, err := snappy.Encode(nil, buf)

			// if err != nil {
			// 	logger.Error("ProtoBufConn snappy.Encode Error %s", err.Error())
			// 	continue
			// }

			conn.c.SetWriteDeadline(time.Now().Add(ConnWriteTimeOut))
			err = binary.Write(conn.c, binary.BigEndian, int32(len(buf)))
			if err != nil {
				//logger.Error("ProtoBufConn Write Error %s", err.Error())
				continue
			}

			conn.c.SetWriteDeadline(time.Now().Add(ConnWriteTimeOut))
			_, err = conn.c.Write(buf)
			if err != nil {
				//logger.Error("ProtoBufConn Write Error %s", err.Error())
				continue
			}
		case <-conn.exit:
			return
		}
	}
}

func (conn *ProtoBufConn) GetRemoteIp() string {
	return conn.c.RemoteAddr().String()
}

func (conn *ProtoBufConn) ReadRequest(req *Request, checkbufsize bool) error {
	var size uint32

	fDebug := func(p int, arg uint32) {
		// 调试fightserver问题
		if checkbufsize {
			return
		}
		// logger.Info("p:", p, "size:", arg)
	}
	conn.c.SetReadDeadline(time.Now().Add(ConnReadTimeOut))
	err := binary.Read(conn.c, binary.BigEndian, &size)

	fDebug(1, size)
	if err != nil {
		return err
	}
	fDebug(2, size)
	if checkbufsize && size > ConnMaxBufSize {
		return errors.New("size > ConnMaxBufSize !!!!")
	}

	if !checkbufsize && size > ConnMaxBufSize {
		conn.readbuf = make([]byte, size)
	}
	fDebug(3, size)
	buf := conn.readbuf[:size]
	conn.c.SetReadDeadline(time.Now().Add(ConnReadTimeOut))

	_, err = io.ReadFull(conn.c, buf)
	if err != nil {
		fDebug(4, size)
		logger.Error("4 error:", err)
		return err
	}

	// dst, err := snappy.Decode(nil, buf)

	// if err != nil {
	// 	fDebug(5, size)
	// 	logger.Error("5 error:", err)
	// 	return err
	// }

	conn.last_time = time.Now().Unix()

	return proto.Unmarshal(buf, req)
}

func (conn *ProtoBufConn) GetRequestBody(req *Request, body interface{}) error {
	if value, ok := body.(proto.Message); ok {
		return proto.Unmarshal(req.GetSerializedRequest(), value)
	}

	return fmt.Errorf("value type error %v", body)
}

func (conn *ProtoBufConn) writeRequest(r *Request) error {
	if conn.is_closed {
		return fmt.Errorf("connection is closed!")
	}

	conn.lockChannelSize.Lock()
	if conn.channelSize > conn.channelMaxSize {
		conn.lockChannelSize.Unlock()
		conn.Close()
		return fmt.Errorf("connection max buf size!")
	}
	conn.channelSize++
	conn.lockChannelSize.Unlock()

	conn.send <- r

	return nil
}

func (conn *ProtoBufConn) Call(serviceMethod string, args interface{}) error {
	var msg proto.Message

	switch m := args.(type) {
	case proto.Message:
		msg = m
	default:
		return fmt.Errorf("Call args type error %v", args)
	}

	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	req := &Request{}
	req.Method = &serviceMethod
	req.SerializedRequest = buf

	return conn.writeRequest(req)
}

func (conn *ProtoBufConn) WriteObj(value interface{}) error {
	var msg proto.Message

	switch m := value.(type) {
	case proto.Message:
		msg = m
	default:
		return fmt.Errorf("WriteObj value type error %v", value)
	}

	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	req := &Request{}

	t := reflect.Indirect(reflect.ValueOf(msg)).Type()
	req.SetMethod(t.PkgPath() + "." + t.Name())
	req.SerializedRequest = buf

	return conn.writeRequest(req)
}

func (conn *ProtoBufConn) SetId(id uint64) {
	conn.id = id
}

func (conn *ProtoBufConn) GetId() uint64 {
	return conn.id
}

func (conn *ProtoBufConn) Close() (errret error) {
	conn.lockForClose.Lock()

	if conn.is_closed {
		conn.lockForClose.Unlock()
		return nil
	}

	if err := conn.c.Close(); err != nil {

		if err := conn.c.SetDeadline(time.Now()); err != nil {
			conn.lockForClose.Unlock()
			return err
		}
		time.Sleep(10 * time.Millisecond)
		if err := conn.c.Close(); err != nil {
			conn.lockForClose.Unlock()
			return err
		}
	}

	if conn.t != nil {
		conn.t.Stop()
	}

	conn.exit <- true
	conn.is_closed = true

	conn.lockForClose.Unlock()

	return nil
}
