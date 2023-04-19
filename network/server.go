package network

import (
	"fmt"
	"time"
)

type ServerOptions struct {
	Transports []Transport
}

type Server struct {
	ServerOptions
	rpcCh  chan RPC
	quitCh chan struct{} // options
}

func NewServer(options ServerOptions) *Server {
	return &Server{
		ServerOptions: options,
		rpcCh:         make(chan RPC),
		quitCh:        make(chan struct{}, 1),
	}
}

func (s *Server) Start() {
	s.InitTransport()
	ticker := time.NewTicker(5 * time.Second)

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			fmt.Printf("%+v\n", rpc)
		case <-s.quitCh:
			break free
		// s.HandleRPC(rpc)
		case <-ticker.C:
			fmt.Println("do stuff every 5 secs tick")
		}
	}

	fmt.Println("server stopped")
}

func (s *Server) InitTransport() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}
