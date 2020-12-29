package gws

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Connection struct {
	wsConnect *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte

	mutex    sync.Mutex
	isClosed bool
}

// 初始化连接
func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	_, _ = WstService.AddrBuckets.Set(wsConn.RemoteAddr().String(), wsConn, time.Second*100)

	fmt.Println(WstService.AddrBuckets.GetAll())

	conn = &Connection{
		wsConnect: wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}

	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()

	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closeed.")
	}
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closeed.")
	}
	return err
}

func (conn *Connection) Close() {
	_, _ = WstService.AddrBuckets.Del(conn.wsConnect.RemoteAddr().String())
	// 线程安全, 可多次调用
	_ = conn.wsConnect.Close()
	// 利用标记让 closeChan 只关闭一次
	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
}

func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			goto ERR
		}
		// 阻塞这里，等待inChan有空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}
	}
ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		deta []byte
		err  error
	)
	for {
		select {
		case deta = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR

		}
		if err = conn.wsConnect.WriteMessage(websocket.BinaryMessage, deta); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
