package network

type NetAddress string

type RPC struct {
	From    NetAddress
	Payload []byte
}

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(NetAddress, []byte) error
	Addr() NetAddress
}
