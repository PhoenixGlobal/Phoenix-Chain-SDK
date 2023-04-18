package state

import (
	"bytes"

	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/ethdb"

	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/trie"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/rlp"
)

// NewStateSync create a new state trie download scheduler.
func NewStateSync(root common.Hash, database ethdb.KeyValueReader, bloom *trie.SyncBloom) *trie.Sync {
	var syncer *trie.Sync
	callback := func(leaf []byte, parent common.Hash) error {
		var obj Account
		if err := rlp.Decode(bytes.NewReader(leaf), &obj); err != nil {
			return err
		}
		syncer.AddSubTrie(obj.Root, 64, parent, nil)
		syncer.AddRawEntry(common.BytesToHash(obj.CodeHash), 64, parent)
		return nil
	}
	syncer = trie.NewSync(root, database, callback, bloom)
	return syncer
}
