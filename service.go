package gws

import (
	"github.com/gorilla/websocket"
	"github.com/yanshuaizhao/gkv"
	"net/http"
)

type WebSocketService struct {
	AddrBuckets *gkv.Gkv
	Config      *Config
}

var (
	WstService *WebSocketService

	upgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// 注册WebSocket服务
func RegisterWsService(c *Config) *WebSocketService {
	WstService = &WebSocketService{
		AddrBuckets: gkv.New(),
		Config:      c,
	}
	return WstService
}

func (s *WebSocketService) Listen(f func(w http.ResponseWriter, r *http.Request)) {
	pattern := s.Config.Pattern
	if pattern == "" {
		pattern = "/ws"
	}
	http.HandleFunc(pattern, f)
	_ = http.ListenAndServe(s.Config.Addr, nil)
}

func (s *WebSocketService) WsHandler(w http.ResponseWriter, r *http.Request, callback func(*Connection) error) {
	var (
		wsConn *websocket.Conn
		err    error
		conn   *Connection
	)
	// header设置
	if wsConn, err = upgrade.Upgrade(w, r, nil); err != nil {
		return
	}

	// 初始化连接
	if conn, err = InitConnection(wsConn); err != nil {
		goto ERR
	}

	for {
		err = callback(conn)
		if err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()

}
