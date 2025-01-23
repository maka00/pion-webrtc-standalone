package signalling

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const bufferSize = 1024

type RouteHandler interface {
	AddRoute(path string, handler func(w http.ResponseWriter, r *http.Request))
}

type HTTPSignallerClient struct {
	routeHandler    RouteHandler
	upgrader        websocket.Upgrader
	websocketClient func(*WebrtcClient)
}

func NewHTTPSignallerClient(r RouteHandler) *HTTPSignallerClient {
	return &HTTPSignallerClient{ //nolint:exhaustruct
		routeHandler: r,
		upgrader: websocket.Upgrader{ //nolint:exhaustruct
			ReadBufferSize:  bufferSize,
			WriteBufferSize: bufferSize,
			CheckOrigin:     func(_ *http.Request) bool { return true },
		},
	}
}

func (sicli *HTTPSignallerClient) OnNewSignallerClient(handler func(client *WebrtcClient)) {
	sicli.websocketClient = handler
}

func (sicli *HTTPSignallerClient) Init() {
	sicli.routeHandler.AddRoute("/ws", sicli.handleWS)
}

func (sicli *HTTPSignallerClient) handleWS(w http.ResponseWriter, r *http.Request) {
	var err error

	var conn *websocket.Conn

	if conn, err = sicli.upgrader.Upgrade(w, r, nil); err != nil {
		log.Fatalf("Error upgrading connection: %v", err)
	}

	sicli.websocketClient(NewWebrtcClient(conn))
}
