# Golang SDK for Phoenix Chain
Developers can easily interact with the Phoenix Chain using this go sdk, examples are as follows.

## Initializing a Client
Before interacting with the blockchain, you need to initialize a client. First import the relevant packages:

```Golang
import (
	phoenixClient "github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/ethclient"
)
```

Then connect a client: 

```Golang
client, err := phoenixClient.Dial("https://dataseed1.phoenix.global/rpc")
```

## Querying Account Balance
Import the relevant packages:

```Golang
import (
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
)
```

Query the balance of an account at the latest block.

```Golang
	account := common.HexToAddress("0xF35F4545fa03416C4ABb9FC9c743EDa6e35c5592")
	latestBalance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
  		log.Fatal(err)
	}
	fmt.Println(latestBalance)
```

You can also query the balance of an account at a certain block height.

```Golang
	blockNumber := big.NewInt(21879393)
	account := common.HexToAddress("0xF35F4545fa03416C4ABb9FC9c743EDa6e35c5592")
	balance, err := client.BalanceAt(context.Background(), account, blockNumber)
	if err != nil {
  		log.Fatal(err)
	}
	fmt.Println(balance)
```

## Querying Blocks
By calling BlockByNumber, you can get all the information of a block, including block height, timestamp, block hash, transaction list, etc.

```Golang
	blockNumber := big.NewInt(21879393)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
  		log.Fatal(err)
	}
	fmt.Println(block.Hash())
	fmt.Println(block.Time())
	fmt.Println(len(block.Transactions()))
```

## Querying Transactions
Loop through the transactions in the block, you can get transaction details.

```Golang
import (
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/core/types"
)
```

```Golang
	for _, trans := range block.Transactions() {
		fmt.Println(trans.Hash().Hex())	// get transaction hash
		fmt.Println(trans.Value().String()) // get transaction value
		chianID, err := client.ChainID(context.Background())
		if err != nil{
			fmt.Println(err)
		}
		msg, err := trans.AsMessage(types.NewEIP155Signer(chianID))
		if err != nil{
			fmt.Println(err)
		} else{
			fmt.Println(msg.From().Hex()) // get from address
		}
		fmt.Println(trans.To().Hex()) // get to address
	}		
```

Or get transaction details directly by transaction hash:

```Golang
	txHash := common.HexToHash("0xdc52533193e5b31707d7a8cacb06eb78acbcd13dfc4917c61a2215806ad7ece6")
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tx.Hash().Hex())
	fmt.Println(isPending)
```

You can also get the transaction receipt by transaction hash.

```Golang
	txHash := common.HexToHash("0xdc52533193e5b31707d7a8cacb06eb78acbcd13dfc4917c61a2215806ad7ece6")
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	fmt.Println(receipt.Status)
	fmt.Println(receipt.Logs)
```

## Smart Contracts
### Install development tools
Before starting work, you need to install the abigen tool.

```shell
# Install libraries
sudo apt install libgmp-dev libssl-dev
# Build to generate abigen
git clone https://github.com/PhoenixGlobal/Phoenix-Chain-SDK.git
cd Phoenix-Chain-SDK
go mod tidy
make all
sudo cp -f ./build/bin/abigen /usr/local/bin/
```

Install the solc compilation tool and prepare the solidity contract file.

```Solidity
pragma solidity ^0.4.24;

contract Store {
    event ItemSet(bytes32 key, bytes32 value);

    string public version;
    mapping (bytes32 => bytes32) public items;

    constructor(string _version) public {
        version = _version;
    }

    function setItem(bytes32 key, bytes32 value) external {
        items[key] = value;
        emit ItemSet(key, value);
    }
}
```

### Generate Go contract file
Generate abi and corresponding golang files.

```Shell
solc --abi --bin Store.sol -o abi
abigen --bin=./abi/Store.bin --abi=./abi/Store.abi --pkg=store --out=Store.go
```

### Deploying a contract
We can copy the generated `Store.go` file to the Golang project, and realize the interaction with the `Store.sol` contract by calling the functions in `Store.go`. Here is a complete example of deploying a contract:

```Golang
package main
import (
	store "...The path of the store.go file..."
	"context"
	"fmt"
	"log"
	"math/big"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/accounts/abi/bind"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/ethclient"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/crypto"
)
func main() {
	client, err := ethclient.Dial("https://dataseed1.phoenix.global/rpc")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA("your account private key")
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	fromAddress := common.HexToAddress("your account address")
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = big.NewInt(200000000000)
	version := "1.0.0"
	contractAddress, tx, contractInstance, err := store.DeployStore(auth, client, version)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contractAddress.Hex()) // print contract address, example: 0x9aC8B849A8b6Fc14F8dEcfa6A22dB41671B38eFB
	fmt.Println(tx.Hash().Hex())       // print transaction has
	_ = contractInstance // new contract instance
}
```

### Querying a contract
After the contract is successfully deployed, you can call querying methods of the contract by the new contract instance. Here is a complete example of querying a contract.

```Golang
package main
import (
	store "...The path of the store.go file..."
	"fmt"
	"log"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/ethclient"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
)
func main() {
	client, err := ethclient.Dial("https://dataseed1.phoenix.global/rpc")
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress("0x9aC8B849A8b6Fc14F8dEcfa6A22dB41671B38eFB")
	contractInstance, err := store.NewStore(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("contract is loaded")
	version, err := contractInstance.Version(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(version) // "1.0.0"
}
```

### Writing to a contract
To calling writing method of the contract, you need to provide the account private key for authentication.

```Golang
package main
import (
	store "...The path of the store.go file..."
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/accounts/abi/bind"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/ethereum/ethclient"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/common"
	"github.com/PhoenixGlobal/Phoenix-Chain-SDK/libs/crypto"
)
func main() {
	client, err := ethclient.Dial("https://dataseed1.phoenix.global/rpc")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA("your private key")
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice
	contractAddress := common.HexToAddress("0x9aC8B849A8b6Fc14F8dEcfa6A22dB41671B38eFB")
	instance, err := store.NewStore(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	key := [32]byte{}
	value := [32]byte{}
	copy(key[:], []byte("testKey"))
	copy(value[:], []byte("testValue"))
	tx, err := instance.SetItem(auth, key, value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("transaction hash:", tx.Hash().Hex())
	result, err := instance.Items(nil, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result[:])) // "testValue"
}
```