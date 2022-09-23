package core 

import (
	"bufio" 
	"encoding/json" 
	"fmt" 
	"os" 
	"reflect" 
	"sort" 

	"github.com/ethereum/go-ethereum/common"
)


const TxGas = 21 
const TxGasPriceDefault = 1 
const TxFee = uint(50) 

type State struct{
	Balances map[common.Address]uint 
	Account2Nonce map[common.Address]uint 

	dbFile *os.File 
	latestBlock Block 
	latestBlockHash Hash 

	miningDifficulty uint  

	forkBCL1 uint64 

	HashCache map[string]uint64 
	HeightCache map[uint64]int64
}


func NewStateFromDisk(dataDir string, miningDifficulty uint)(*State, error){
	err := InitialDirIfNotExists(dataDir, []byte(genesisJson))
}