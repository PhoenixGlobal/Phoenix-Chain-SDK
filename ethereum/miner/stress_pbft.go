// +build none

// This file contains a miner stress test based on the pbft consensus engine.
package main

import (
	eth2 "github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/eth"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/eth/downloader"
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/configs"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/accounts/keystore"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/core"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/core/types"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/node"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/p2p"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/p2p/discover"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common/fdlimit"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/crypto"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/log"
)

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	fdlimit.Raise(2048)

	// Generate a batch of accounts to seal and fund with
	faucets := make([]*ecdsa.PrivateKey, 128)
	for i := 0; i < len(faucets); i++ {
		faucets[i], _ = crypto.GenerateKey()
	}
	sealers := make([]*ecdsa.PrivateKey, 1)
	for i := 0; i < len(sealers); i++ {
		sealers[i], _ = crypto.GenerateKey()
	}
	// Create a Clique network based off of the Rinkeby config
	genesis := makeGenesis(faucets, sealers)

	var (
		nodes  []*node.Node
		enodes []string
	)
	for _, sealer := range sealers {
		// Start the node and wait until it's up
		node, err := makeSealer(genesis, enodes)
		if err != nil {
			panic(err)
		}
		defer node.Close()

		for node.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to al the previous ones
		for _, enode := range enodes {
			enode, err := discover.ParseNode(enode)
			if err != nil {
				panic(err)
			}
			node.Server().AddPeer(enode)
		}
		// Start tracking the node and it's enode url
		nodes = append(nodes, node)

		enode := fmt.Sprintf("enode://%s@127.0.0.1:%d", node.Server().NodeInfo().ID, node.Server().NodeInfo().Ports.Listener)
		enodes = append(enodes, enode)

		// Inject the signer key and start sealing with it
		store := node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
		signer, err := store.ImportECDSA(sealer, "")
		if err != nil {
			panic(err)
		}
		if err := store.Unlock(signer, ""); err != nil {
			panic(err)
		}
	}
	// Iterate over all the nodes and start signing with them
	time.Sleep(3 * time.Second)

	for _, node := range nodes {
		var ethereum *eth2.Ethereum
		if err := node.Service(&ethereum); err != nil {
			panic(err)
		}
	}
	time.Sleep(3 * time.Second)

	// Start injecting transactions from the faucet like crazy
	nonces := make([]uint64, len(faucets))
	for {
		index := rand.Intn(len(faucets))

		// Fetch the accessor for the relevant signer
		var ethereum *eth2.Ethereum
		if err := nodes[index%len(nodes)].Service(&ethereum); err != nil {
			panic(err)
		}
		// Create a self transaction and inject into the pool
		tx, err := types.SignTx(types.NewTransaction(nonces[index], crypto.PubkeyToAddress(faucets[index].PublicKey), new(big.Int), 21000, big.NewInt(100000000000), nil), types.NewEIP155Signer(new(big.Int)), faucets[index])
		if err != nil {
			panic(err)
		}
		if err := ethereum.TxPool().AddLocal(tx); err != nil {
			panic(err)
		}
		nonces[index]++

		// Wait if we're too saturated
		if pend, _ := ethereum.TxPool().Stats(); pend > 2048 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// makeGenesis creates a custom Clique genesis block based on some pre-defined
// signer and faucet accounts.
func makeGenesis(faucets []*ecdsa.PrivateKey, sealers []*ecdsa.PrivateKey) *core.Genesis {
	// Create a Clique network based off of the Rinkeby config

	genesis := core.DefaultGrapeGenesisBlock()
	genesis.GasLimit = 3150000000

	genesis.Config.ChainID = big.NewInt(304)
	genesis.Config.Pbft.Duration = 10

	genesis.Alloc = core.GenesisAlloc{}
	for _, faucet := range faucets {
		genesis.Alloc[crypto.PubkeyToAddress(faucet.PublicKey)] = core.GenesisAccount{
			Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		}
	}
	// Sort the signers and embed into the extra-data section
	signers := make([]common.Address, len(sealers))
	for i, sealer := range sealers {
		signers[i] = crypto.PubkeyToAddress(sealer.PublicKey)
	}
	for i := 0; i < len(signers); i++ {
		for j := i + 1; j < len(signers); j++ {
			if bytes.Compare(signers[i][:], signers[j][:]) > 0 {
				signers[i], signers[j] = signers[j], signers[i]
			}
		}
	}
	genesis.ExtraData = make([]byte, 32+len(signers)*common.AddressLength+65)
	for i, signer := range signers {
		copy(genesis.ExtraData[32+i*common.AddressLength:], signer[:])
	}
	// Return the genesis block for initialization
	return genesis
}

func makeSealer(genesis *core.Genesis, nodes []string) (*node.Node, error) {
	// Define the basic configurations for the Ethereum node
	datadir, _ := ioutil.TempDir("", "")

	config := &node.Config{
		Name:    "phoenixchain",
		Version: configs.Version,
		DataDir: datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
		NoUSB:       true,
		KeyStoreDir: "D:\\goprojects\\data\\keystore",
	}
	// Start the node and configure a full Ethereum node on it
	stack, err := node.New(config)
	if err != nil {
		return nil, err
	}
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return eth2.New(ctx, &eth2.Config{
			Genesis:         genesis,
			NetworkId:       genesis.Config.ChainID.Uint64(),
			SyncMode:        downloader.FullSync,
			DatabaseCache:   256,
			DatabaseHandles: 256,
			TxPool:          core.DefaultTxPoolConfig,
			GPO:             eth2.DefaultConfig.GPO,
			MinerGasFloor:   genesis.GasLimit * 9 / 10,
			MinerGasCeil:    genesis.GasLimit * 21 / 10,
			MinerGasPrice:   big.NewInt(1),
			MinerRecommit:   time.Second,
		})
	}); err != nil {
		return nil, err
	}
	// Start the node and return if successful
	return stack, stack.Start()
}
