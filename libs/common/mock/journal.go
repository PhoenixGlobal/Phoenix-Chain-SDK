package mock

import (
	"golang.org/x/crypto/sha3"
	"math/big"

	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/rlp"

	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
)

type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(*MockStateDB)
}

type journal struct {
	entries []journalEntry // Current changes tracked by the journal
}

// newJournal create a new initialized journal.
func NewJournal() *journal {
	return &journal{
		entries: []journalEntry{},
	}
}

func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
}

func (j *journal) revert(statedb *MockStateDB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		// Undo the changes made by the operation
		j.entries[i].revert(statedb)
	}
	j.entries = j.entries[:snapshot]
}

// length returns the current number of entries in the journal.
func (j *journal) length() int {
	return len(j.entries)
}

type (
	balanceChange struct {
		account *common.Address
		prev    *big.Int
		newOne  bool
	}

	storageChange struct {
		account  *common.Address
		key      []byte
		preValue []byte
	}

	createObjectChange struct {
		account *common.Address
	}

	nonceChange struct {
		account *common.Address
		prev    uint64
		newOne  bool
	}

	codeChange struct {
		account  *common.Address
		prevcode []byte
		newOne   bool
	}

	suicideChange struct {
		account     *common.Address
		prevbalance *big.Int
	}

	addLogChange struct {
		txhash common.Hash
	}
)

func (ch balanceChange) revert(s *MockStateDB) {
	s.Balance[*ch.account] = ch.prev
	if ch.newOne {
		delete(s.Balance, *ch.account)
	}
}

func (ch storageChange) revert(s *MockStateDB) {
	if len(ch.preValue) == 0 {
		delete(s.State[*ch.account], string(ch.key))
	} else {
		s.State[*ch.account][string(ch.key)] = ch.preValue
	}
}

func (ch createObjectChange) revert(s *MockStateDB) {
	delete(s.State, *ch.account)
}

func (ch nonceChange) revert(s *MockStateDB) {
	if ch.newOne {
		delete(s.Nonce, *ch.account)
	} else {
		s.Nonce[*ch.account] = ch.prev
	}
}

func (ch codeChange) revert(s *MockStateDB) {
	if ch.newOne {
		delete(s.Code, *ch.account)
		delete(s.CodeHash, *ch.account)
	} else {
		s.Code[*ch.account] = ch.prevcode

		var h common.Hash
		hw := sha3.NewLegacyKeccak256()
		rlp.Encode(hw, ch.prevcode)
		hw.Sum(h[:0])
		s.CodeHash[*ch.account] = h[:]
	}
}

func (ch suicideChange) revert(s *MockStateDB) {
	delete(s.Suicided, *ch.account)
	s.Balance[*ch.account] = ch.prevbalance
}

func (ch addLogChange) revert(s *MockStateDB) {
	logs := s.Logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.Logs, ch.txhash)
	} else {
		s.Logs[ch.txhash] = logs[:len(logs)-1]
	}
	s.logSize--
}
