package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/dbkbali/bcbasic/core"
	"github.com/sirupsen/logrus"
)

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
)

type RPC struct {
	From    NetAddress
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{
		Header: t,
		Data:   data,
	}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(msg)
	return buf.Bytes()
}

type DecodeMessage struct {
	From NetAddress
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodeMessage, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodeMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Header,
	}).Debug("new incoming message")

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: tx,
		}, nil

	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}

		return &DecodeMessage{
			From: rpc.From,
			Data: block,
		}, nil

	default:
		return nil, fmt.Errorf("unknown message header type: %x", msg.Header)
	}
}

type RPCProcessor interface {
	ProcessMessage(*DecodeMessage) error
}
