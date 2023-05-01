package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dbkbali/bcbasic/core"
	"github.com/dbkbali/bcbasic/crypto"
	"github.com/dbkbali/bcbasic/types"
	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOptions struct {
	SeedNodes     []string
	ListenAddr    string
	TCPTransport  *TCPTransport
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	TCPTransport *TCPTransport
	peerCh       chan (*TCPPeer)

	mu      sync.RWMutex
	peerMap map[net.Addr]*TCPPeer

	ServerOptions
	memPool     *TxPool
	chain       *core.Blockchain
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{} // options
}

func NewServer(options ServerOptions) (*Server, error) {
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}
	if options.RPCDecodeFunc == nil {
		options.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if options.Logger == nil {
		options.Logger = log.NewLogfmtLogger(os.Stderr)
		options.Logger = log.With(options.Logger, "addr", options.ID)
	}

	chain, err := core.NewBlockchain(options.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	// apiServer :=
	peerCh := make(chan *TCPPeer)
	tr := NewTCPTransport(options.ListenAddr, peerCh)
	s := &Server{
		TCPTransport:  tr,
		peerCh:        peerCh,
		peerMap:       make(map[net.Addr]*TCPPeer),
		ServerOptions: options,
		chain:         chain,
		memPool:       NewTxPool(1000),
		isValidator:   options.PrivateKey != nil,
		rpcCh:         make(chan RPC),
		quitCh:        make(chan struct{}, 1),
	}

	s.TCPTransport.peerCh = peerCh

	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) bootstrapNetwork() {
	for _, addr := range s.SeedNodes {

		go func(addr string) {

			conn, err := net.Dial("tcp", addr)
			if err != nil {
				s.Logger.Log("err", err)
				return
			}

			peer := &TCPPeer{
				conn: conn,
			}

			s.peerCh <- peer

		}(addr)

	}
}

func (s *Server) Start() {
	s.TCPTransport.Start()

	time.Sleep(time.Second * 2)

	s.bootstrapNetwork()

	s.Logger.Log("accepting tcp on", s.ListenAddr, "id", s.ID)

free:
	for {
		select {
		case peer := <-s.peerCh:
			s.peerMap[peer.conn.RemoteAddr()] = peer

			go peer.readLoop(s.rpcCh)

			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}

			s.Logger.Log("msg", "new peer added", "outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())

		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("err", err)
				continue
			}

			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					s.Logger.Log("err", err)
				}
			}
		case <-s.quitCh:
			break free
		}
	}

	s.Logger.Log("msg", "server stopped")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Log("msg", "Starting validator loop", "blockTime", s.BlockTime)

	for {
		<-ticker.C
		s.CreateNewBlock()
	}
}

func (s *Server) ProcessMessage(msg *DecodeMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From, t)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, t)

	}

	return nil
}

// GGG
func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	s.Logger.Log("msg", "received GET BLOCKS msg", "from", from)

	var (
		blocks    = []*core.Block{}
		ourHeight = s.chain.Height()
	)

	if data.To == 0 {
		for i := int(data.From); i <= int(ourHeight); i++ {
			block, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, block)
		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer not found")
	}

	return peer.Send(msg.Bytes())
}

func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for netAddr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			s.Logger.Log("err", err, "addr", netAddr)
		}

	}
	return nil
}

func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	s.Logger.Log("msg", "received BLOCKS !!!!!", "from", from)

	for _, block := range data.Blocks {
		if err := s.chain.AddBlock(block); err != nil {
			fmt.Printf("BLOCK ERROR: %s\n", err)
			continue
			// s.Logger.Log("err", err)
		}
	}

	return nil
}

func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	fmt.Printf("received STATUS msg from: [%s] => %+v\n", from, data)
	if data.CurrentHeight <= s.chain.Height() {
		s.Logger.Log("msg", "cannot sync remote node blockheight to low", "our height", s.chain.Height(), "height", data.CurrentHeight, "addr", from)
		return nil
	}

	// this remote has blocks we can sync
	go s.requestBlocksLoop(from)
	return nil
}

func (s *Server) processGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
	s.Logger.Log("msg", "received GET STATUS msg from", "addr", from)

	statusMsg := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.ID,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMsg); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer not found")
	}

	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	go s.broadcastBlock(b)

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	// s.Logger.Log(
	// 	"msg", "added new tx to pool",
	// 	"hash", hash,
	// 	"mempool len", s.memPool.PendingCount(),
	// )

	go s.broadcastTx(tx)

	s.memPool.Add(tx)
	return nil
}

// TODO: stop syncing when at highest block
func (s *Server) requestBlocksLoop(peer net.Addr) error {
	ticker := time.NewTicker(3 * time.Second)

	for {
		ourHeight := s.chain.Height()
		s.Logger.Log("msg", "requesting blocks", "requesting height", ourHeight+1, "addr", peer)
		// this remote has blocks we can sync
		getBlocksMsg := &GetBlocksMessage{
			From: ourHeight + 1,
			To:   0,
		}

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlocksMsg); err != nil {
			return err
		}

		s.mu.RLock()
		defer s.mu.RUnlock()

		msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
		peer, ok := s.peerMap[peer]
		if !ok {
			return fmt.Errorf("peer %+s not found", peer.conn.RemoteAddr())
		}

		if err := peer.Send(msg.Bytes()); err != nil {
			s.Logger.Log("error", "failed to send to peer", "err", err, "peer", peer.conn.RemoteAddr())
		}

		<-ticker.C
	}
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) CreateNewBlock() error {
	// 1. get transactions from mempool
	// 2. create a new block
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// TODO: change from adding all txs to pool - limit via some function later
	// To match the tx types
	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	s.memPool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}

	b, _ := core.NewBlock(header, nil)

	privKey := crypto.GeneratePrivateKey()
	if err := b.Sign(privKey); err != nil {
		panic(err)
	}

	return b
}
