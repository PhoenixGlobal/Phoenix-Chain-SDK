package types

import (
	"bytes"

	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/trie"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/rlp"
)

type DerivableList interface {
	Len() int
	GetRlp(i int) []byte
}

func DeriveSha(list DerivableList) common.Hash {
	keybuf := new(bytes.Buffer)
	trie := new(trie.Trie)
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRlp(i))
	}
	return trie.Hash()
}
