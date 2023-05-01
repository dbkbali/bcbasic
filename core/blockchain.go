package core

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type Blockchain struct {
	logger    log.Logger
	store     Storage
	lock      sync.RWMutex
	headers   []*Header
	blocks    []*Block
	validator Validator
	// TODO: convert to interface
	contractState *State
}

func NewBlockchain(l log.Logger, genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		contractState: NewState(),
		headers:       []*Header{},
		store:         NewMemoryStore(),
		logger:        l,
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(genesis)

	return bc, err
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(b *Block) error {
	// validate block
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}

	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "running vm", "len", len(tx.Data), "hash", tx.Hash(TxHasher{}))

		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}

		// vmResult := vm.stack.Pop()

	}

	return bc.addBlockWithoutValidation(b)
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("blockchain height [%d] is less than requested height [%d]", bc.Height(), height)
	}
	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.blocks[height], nil
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("blockchain height [%d] is less than requested height [%d]", bc.Height(), height)
	}
	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.headers[height], nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	bc.lock.Unlock()

	bc.logger.Log(
		"msg", "BlockAdd",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Height,
		"txs", len(b.Transactions),
	)
	return bc.store.Put(b)
}
