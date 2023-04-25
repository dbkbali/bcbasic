package main

import (
	"bytes"
	"log"
	"math/rand"

	"strconv"
	"time"

	"github.com/dbkbali/bcbasic/core"
	"github.com/dbkbali/bcbasic/crypto"
	"github.com/dbkbali/bcbasic/network"
	"github.com/sirupsen/logrus"
)

func main() {
	trLocal := network.NewLocalTransport("local")
	trRemote := network.NewLocalTransport("remote")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			// trRemote.SendMessage(trLocal.Addr(), []byte("hello world"))
			sendTransaction(trLocal, trRemote.Addr())
			if err := sendTransaction(trRemote, trLocal.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	privKey := crypto.GeneratePrivateKey()
	opts := network.ServerOptions{
		PrivateKey: &privKey,
		ID:         "LocalServer",
		Transports: []network.Transport{trLocal},
	}

	srv, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	srv.Start()
}

func sendTransaction(tr network.Transport, to network.NetAddress) error {
	privKey := crypto.GeneratePrivateKey()

	data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())

}
