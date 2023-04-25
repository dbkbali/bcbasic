package network

import (
	"bytes"
	"crypto"
	"fmt"
	"time"

	"github.com/dbkbali/bcbasic/core"
	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOptions
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{} // options
}

func NewServer(options ServerOptions) *Server {
	if options.BlockTime == 0 {
		options.BlockTime = defaultBlockTime
	}
	if options.RPCDecodeFunc == nil {
		options.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	s := &Server{
		ServerOptions: options,
		memPool:       NewTxPool(),
		isValidator:   options.PrivateKey != nil,
		rpcCh:         make(chan RPC),
		quitCh:        make(chan struct{}, 1),
	}

	// if config doesn't designate a processor, use the server

	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	return s
}

func (s *Server) Start() {
	s.InitTransport()
	ticker := time.NewTicker(s.BlockTime)

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				logrus.Error(err)
			}

			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				logrus.Error(err)
			}
		case <-s.quitCh:
			break free
		// s.HandleRPC(rpc)
		case <-ticker.C:
			if s.isValidator {
				// need consensus logic here
				s.CreateNewBlock()
			}
		}
	}

	fmt.Println("server stopped")
}

func (s *Server) CreateNewBlock() error {
	// 1. get transactions from mempool
	// 2. create a new block
	fmt.Println("creating a new block")
	return nil
}

func (s *Server) ProcessMessage(msg *DecodeMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	}

	return nil
}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash":           hash,
			"mempool length": s.memPool.Len(),
		}).Info("tx already exists in mempool")

		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	logrus.WithFields(logrus.Fields{
		"hash": tx.Hash(core.TxHasher{})}).Info("adding new tx to mempool")

	go s.broadcastTx(tx)

	return s.memPool.Add(tx)
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
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
