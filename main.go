package main

import (
	"bytes"
	"log"
	"time"

	"strconv"

	"github.com/dbkbali/bcbasic/core"
	"github.com/dbkbali/bcbasic/crypto"
	"github.com/dbkbali/bcbasic/network"
	"github.com/sirupsen/logrus"
)

func main() {
	trLocal := network.NewLocalTransport("local")
	trRemoteA := network.NewLocalTransport("remote-a")
	trRemoteB := network.NewLocalTransport("remote-b")
	trRemoteC := network.NewLocalTransport("remote-c")

	trLocal.Connect(trRemoteA)
	trRemoteA.Connect(trRemoteB)
	trRemoteB.Connect(trRemoteC)

	trRemoteA.Connect(trLocal)

	initRemoteServers([]network.Transport{trRemoteA, trRemoteB, trRemoteC})

	go func() {
		for {
			// trRemote.SendMessage(trLocal.Addr(), []byte("hello world"))
			sendTransaction(trRemoteA, trLocal.Addr())
			if err := sendTransaction(trRemoteA, trLocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	// go func() {
	// 	time.Sleep(7 * time.Second)

	// 	trRemoteX := network.NewLocalTransport("remote-x")
	// 	trRemoteC.Connect(trRemoteX)
	// 	lateServer := makeServer("Late", trRemoteX, nil)

	// 	go lateServer.Start()
	// }()

	privKey := crypto.GeneratePrivateKey()

	srvLocal := makeServer("Local1", trLocal, &privKey)
	srvLocal.Start()

}

func initRemoteServers(trs []network.Transport) {
	for i, tr := range trs {
		id := "Remote" + strconv.Itoa(i)
		srv := makeServer(id, tr, nil)
		go srv.Start()
	}
}

func makeServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
	options := network.ServerOptions{
		PrivateKey: pk,
		ID:         id,
		Transports: []network.Transport{tr},
	}

	s, err := network.NewServer(options)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func sendTransaction(tr network.Transport, to network.NetAddress) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())

}
