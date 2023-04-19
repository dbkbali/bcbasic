package main

import (
	"time"

	"github.com/dbkbali/bcbasic/network"
)

func main() {
	trLocal := network.NewLocalTransport("local")
	trRemote := network.NewLocalTransport("remote")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			trRemote.SendMessage(trLocal.Addr(), []byte("hello world"))
			time.Sleep(1 * time.Second)
		}
	}()

	opts := network.ServerOptions{
		Transports: []network.Transport{trLocal},
	}

	srv := network.NewServer(opts)
	srv.Start()
}
