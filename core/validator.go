package core

import "fmt"

type Validator interface {
	ValidateBlock(b *Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	// validate block header
	// validate block transactions
	// validate block hash
	if v.bc.HasBlock(b.Height) {
		return fmt.Errorf("block [%d] already exists with hash [%x]", b.Height, b.Hash(BlockHasher{}))
	}

	if b.Height != v.bc.Height()+1 {
		return fmt.Errorf("block height [%d] is too high height [%d] - block %s", b.Height, v.bc.Height()+1, b.Hash(BlockHasher{}))
	}

	prevHeader, err := v.bc.GetHeader(b.Height - 1)
	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.PrevBlockHash {
		return fmt.Errorf("block prev hash [%x] does not match prev header hash [%x]", b.PrevBlockHash, hash)
	}

	if err := b.Verify(); err != nil {
		return err
	}

	return nil
}
