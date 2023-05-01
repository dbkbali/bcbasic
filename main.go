package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"time"

	"github.com/dbkbali/bcbasic/core"
	"github.com/dbkbali/bcbasic/crypto"
	"github.com/dbkbali/bcbasic/network"
)

// var transports = []network.Transport{
// 	network.NewLocalTransport("local"),
// 	network.NewLocalTransport("remote-a"),
// 	//network.NewLocalTransport("remote-b"),
// 	//network.NewLocalTransport("remote-c"),
// 	// network.NewLocalTransport("late"),
// }

func main() {
	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":4000"})

	go localNode.Start()

	remoteNode := makeServer("REMOTE_A", nil, ":4000", []string{":5000"})
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_B", nil, ":5000", nil)
	go remoteNodeB.Start()

	go func() {
		time.Sleep(11 * time.Second)
		lateNode := makeServer("LATE_NODE", nil, ":6000", []string{":4000"})
		go lateNode.Start()

	}()

	time.Sleep(1 * time.Second)

	// tcpTester()

	select {}
}

func makeServer(id string, pk *crypto.PrivateKey, addr string, seedNodes []string) *network.Server {
	options := network.ServerOptions{
		SeedNodes:  seedNodes,
		ListenAddr: addr,
		PrivateKey: pk,
		ID:         id,
	}

	s, err := network.NewServer(options)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

// func main() {
// 	initRemoteServers(transports)
// 	localNode := transports[0]
// 	lateTr := network.NewLocalTransport("LATE")
// 	// remoteNodeA := transports[1]

// 	// go func() {
// 	// 	for {
// 	// 		if err := sendTransaction(remoteNodeA, localNode.Addr()); err != nil {
// 	// 			logrus.Error(err)
// 	// 		}
// 	// 		time.Sleep(2 * time.Second)
// 	// 	}
// 	// }()

// 	// trLate := network.NewLocalTransport("late-remote")
// 	go func() {
// 		time.Sleep(7 * time.Second)
// 		lateServer := makeServer(string(lateTr.Addr()), lateTr, nil)
// 		go lateServer.Start()
// 	}()

// 	privKey := crypto.GeneratePrivateKey()

// 	srvLocal := makeServer("Local1", localNode, &privKey)

// 	srvLocal.Start()

// }

func initRemoteServers(trs []network.Transport) {
	// for i, tr := range trs {
	// 	id := "Remote" + strconv.Itoa(i)
	// 	srv := makeServer(id, nil)
	// 	go srv.Start()
	// }
}

func sendGetStatusMessage(tr network.Transport, to net.Addr) error {
	var (
		getStatusMsg = new(network.GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())
}

func sendTransaction(tr network.Transport, to net.Addr) error {
	data := contract()
	privKey := crypto.GeneratePrivateKey()
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return tr.SendMessage(to, msg.Bytes())

}

func contract() []byte {
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data = append(data, pushFoo...)

	return data
}
