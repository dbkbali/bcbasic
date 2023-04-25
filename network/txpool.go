package network

import (
	"sort"

	"github.com/dbkbali/bcbasic/core"
	"github.com/dbkbali/bcbasic/types"
)

type TxPool struct {
	transactions map[types.Hash]*core.Transaction
}

type TxMapSorter struct {
	transactions []*core.Transaction
}

func NewTxMapSorter(txm map[types.Hash]*core.Transaction) *TxMapSorter {
	txs := make([]*core.Transaction, len(txm))

	i := 0
	for _, val := range txm {
		txs[i] = val
		i++
	}

	s := &TxMapSorter{
		transactions: txs,
	}

	sort.Sort(s)

	return s

}

func (s *TxMapSorter) Len() int { return len(s.transactions) }
func (s *TxMapSorter) Swap(i, j int) {
	s.transactions[i], s.transactions[j] = s.transactions[j], s.transactions[i]
}
func (s *TxMapSorter) Less(i, j int) bool {
	return s.transactions[i].FirstSeen() < s.transactions[j].FirstSeen()
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[types.Hash]*core.Transaction),
	}
}

func (p *TxPool) Transactions() []*core.Transaction {
	s := NewTxMapSorter(p.transactions)
	return s.transactions
}

// Add adds a transaction to the pool. Caller is responsible for
// checking if the tx doesn't already exist in the pool.
func (p *TxPool) Add(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	p.transactions[hash] = tx

	return nil
}

func (p *TxPool) Has(hash types.Hash) bool {
	_, ok := p.transactions[hash]
	return ok
}

func (p *TxPool) Len() int {
	return len(p.transactions)
}

func (p *TxPool) Flush() {
	p.transactions = make(map[types.Hash]*core.Transaction)
}
