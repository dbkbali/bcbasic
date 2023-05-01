package network

import "github.com/dbkbali/bcbasic/core"

type GetBlocksMessage struct {
	From uint32

	// To = 0 max blocks returned
	To uint32
}

type BlocksMessage struct {
	Blocks []*core.Block
}

type GetStatusMessage struct {
}

type StatusMessage struct {
	ID            string
	Version       uint32
	CurrentHeight uint32
}
