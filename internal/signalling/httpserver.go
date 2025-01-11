package signalling

import (
	"context"
	"log"
	"net/http"
	"time"
)

type HTTPServer struct {
	srv         *http.Server
	route       *http.ServeMux
	pathToServe string
}

func NewHTTPServer(rootDir string) *HTTPServer {
	return &HTTPServer{
		srv: &http.Server{ //nolint:exhaustruct
			ReadHeaderTimeout: time.Second,
		},
		route:       &http.ServeMux{},
		pathToServe: rootDir,
	}
}

func (httpServer *HTTPServer) AddRoute(route string, handler func(w http.ResponseWriter, r *http.Request)) {
	httpServer.route.HandleFunc(route, handler)
}

func (httpServer *HTTPServer) Start() {
	httpServer.route.HandleFunc("/", http.FileServer(http.Dir(httpServer.pathToServe)).ServeHTTP)
	httpServer.srv.Handler = httpServer.route
	httpServer.srv.Addr = ":8080"

	go func() {
		if err := httpServer.srv.ListenAndServe(); err != nil {
			log.Printf("error starting http server: %v\n", err.Error())
		}
	}()
}

func (httpServer *HTTPServer) Stop() {
	if err := httpServer.srv.Shutdown(context.Background()); err != nil {
		log.Printf("Error stopping httpserver: %v\n", err.Error())
	}
}
