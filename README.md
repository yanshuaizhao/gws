#### 简介
GWS - Golang WebSocket的缩写，基于gorilla/websocket开发的用于学习研究简单框架 (会持续完善及功能增强)

#### 应用实例
```
package main

import (
	"github.com/yanshuaizhao/gws"
	"net/http"
)

var wsService *gws.WebSocketService

func handler(resp http.ResponseWriter, req *http.Request) {
	wsService.WsHandler(resp, req, func(conn *gws.Connection) error {
		data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		// 业务处理...

		if err := conn.WriteMessage(data); err != nil {
			return err
		}
		return nil
	})
}

func main() {
	wsService = gws.RegisterWsService(&gws.Config{Addr: "127.0.0.1:9199"})
	wsService.Listen(handler)
}
```