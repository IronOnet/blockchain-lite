package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/irononet/blockchain-lite/utils/file/fs"
)

const (
	TxGas = 21 
	TxGasPriceDefault = 1
	TxFee = uint(50)
)

type State struct{
	
}

func GetNextBlock(blockHash Hash, dbDir string) ([]Block, error){
	f , err := os.OpenFile(fs.getBlocksDBFilePath(dbDir), os.O_RDONLY, 0600) 
	if err != nil{
		return nil, err 
	}

	blocks := make([]Block, 0) 
	shouldStartCollecting := false  

	if reflect.DeepEqual(blockHash, Hash{}){
		shouldStartCollecting = true 
	}


	scanner := bufio.NewScanner(f) 
	for scanner.Scan(){
		if err := scanner.Err(); err != nil{
			return nil, err 
		}
		var blockFS BlockFS 
		err = json.Unmarshal(scanner.Bytes(), &blockFS)
		if err != nil{
			return nil, err 
		}

		if shouldStartCollecting{
			blocks = append(blocks, blockFS.Value) 
			continue 
		}

		if blockHash == blockFS.Key{
			shouldStartCollecting = true 
		}
	}
	return blocks, nil 
}


// GetBlockByHeightOrHash returns the desired block by hash or height 
// it uses the cached data in the state struct 
func GetBlockByHeightOrHash(state *State, height uint64, hash, dataDir string) (BlockFS, error){

	var block BlockFS 

	key, ok := state.HeightCache[height] 
	if hash != ""{
		key, ok = state.HeightCache[hash]
	}

	if !ok{
		if hash != ""{
			return block, fmt.Errorf("invalid hash: '%v'", hash)
		}

		return block, fmt.Errorf("invalid height: '%v'", height) 
	}

	f, err := os.OpenFile(fs.getBlocksDBFilePath(dataDir), os.O_RDONLY, 0600)
	if err != nil{
		return block, err 
	}

	scanner := bufio.NewScanner(f) 
	if scanner.Scan(){
		if err := scanner.Err(); err != nil{
			return block , err 
		}

		err = json.Unmarshal(scanner.Bytes(), &block) 
		if err != nil{
			return block, err
		}
	}
	return block, nil 
}