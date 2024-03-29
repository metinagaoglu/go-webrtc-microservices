package websocket

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"log"
	"net/http"

	wsutil "github.com/gobwas/ws/wsutil"
	epoll "websocket-gateway/internal/epoll"
	handlers "websocket-gateway/internal/handlers"
	middleware "websocket-gateway/internal/middleware"
	utils "websocket-gateway/pkg/utils"
)

func WsHandler(w http.ResponseWriter, r *http.Request) {
	epoller := epoll.GetEpollInstance()

	userId, err := Auth(r)
	if err != nil {
		log.Printf("Failed to auth %v", err)
		return
	}

	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	nodeId := utils.GetNodeId()

	// Add context to conn
	ctx := context.Background()
	ctx = context.WithValue(ctx, "userId", userId)
	ctx = context.WithValue(ctx, "nodeId", nodeId)

	// Post connection middlewares
	middleware.InitSessionMiddleware(ctx, conn)

	if err := epoller.Add(conn, ctx); err != nil {
		log.Printf("Failed to add connection %v", err)
		conn.Close()
	}
}

func Start() {
	epoller := epoll.GetEpollInstance()

	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if msg, _, err := wsutil.ReadClientData(conn); err != nil {
				fmt.Println("Disconnect handler", conn)
				if err := epoller.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
				}
				ctx := epoller.GetContext(conn)
				//TODO: close session
				middleware.EndSessionMiddleware(ctx, conn)
				conn.Close()
			} else {
				// This is commented out since in demo usage, stdout is showing messages sent from > 1M connections at very high rate
				log.Printf("msg: %s", string(msg))

				ctx := epoller.GetContext(conn)
				//ctx.WithValue(ctx, "nodeId", nodeId)

				// All websocket messages to be handled here
				go handlers.Run(&conn, ctx, msg)

				err = wsutil.WriteServerMessage(conn, 1, msg)
				if err != nil {
					//TODO: handle error
				}
			}
		}
	}
}
