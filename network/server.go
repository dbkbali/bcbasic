package network

import (
	"crypto"
	"fmt"
	"time"

	"github.com/dbkbali/bcbasic/core"
	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	RPCHandler RPCHandler
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	ServerOptions
	blockTime   time.Duration
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{} // options
}

func NewServer(options ServerOptions) *Server {
	if options.BlockTime == 0 {
		options.BlockTime = defaultBlockTime
	}
	s := &Server{
		ServerOptions: options,
		blockTime:     options.BlockTime,
		memPool:       NewTxPool(),
		isValidator:   options.PrivateKey != nil,
		rpcCh:         make(chan RPC),
		quitCh:        make(chan struct{}, 1),
	}
	if options.RPCHandler == nil {
		options.RPCHandler = NewDefaultRPCHandler(s)
	}

	s.ServerOptions = options

	return s
}

func (s *Server) Start() {
	s.InitTransport()
	ticker := time.NewTicker(s.BlockTime)

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			if err := s.RPCHandler.HandleRPC(rpc); err != nil {
				logrus.WithError(err).Error("error handling rpc")
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

func (s *Server) ProcessTransaction(from NetAddress, tx *core.Transaction) error {
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

	// TDOD: broadcast tx to other nodes

	return s.memPool.Add(tx)
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
