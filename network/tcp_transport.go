package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type TCPPeer struct {
	conn     net.Conn
	Outgoing bool
}

func (p *TCPPeer) Send(payload []byte) error {
	_, err := p.conn.Write(payload)
	return err
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection from [%+v]\n", conn)
			continue
		}

		peer := &TCPPeer{
			conn: conn,
		}

		t.peerCh <- peer

		fmt.Printf("Accepted connection from [%+v]\n", peer.conn)

	}

}

func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 4096)
	for {
		n, err := p.conn.Read(buf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Printf("Error reading from [%+v]\n", p.conn)
			continue
		}

		msg := buf[:n]
		rpcCh <- RPC{
			From:    p.conn.RemoteAddr(),
			Payload: bytes.NewReader(msg),
		}
	}
}

type TCPTransport struct {
	peerCh     chan *TCPPeer
	listenAddr string
	listener   net.Listener
}

func NewTCPTransport(addr string, peerCh chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		peerCh:     peerCh,
		listenAddr: addr,
	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	go t.acceptLoop()

	fmt.Println("TCP TRANSPORT: Listening on port:", t.listenAddr)

	return nil
}
