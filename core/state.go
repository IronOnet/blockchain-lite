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
	hasGenesisBlock bool 

	miningDifficulty uint  

	forkBCL1 uint64 

	HashCache map[string]int64 
	HeightCache map[uint64]int64
}


func NewStateFromDisk(dataDir string, miningDifficulty uint)(*State, error){
	err := InitialDirIfNotExists(dataDir, []byte(genesisJson))
	if err != nil{
		return nil, err 
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir)) 
	if err != nil{
		return nil, err 
	}

	balances := make(map[common.Address]uint) 
	for account, balance := range gen.Balances{
		balances[account] = balance; 
	}

	account2nonce := make(map[common.Address]uint) 

	dbFilePath := getBlocksDBFilePath(dataDir)
	f, err := os.OpenFile(dbFilePath, os.O_APPEND, 0600) 
	if err != nil{
		return nil, err 
	}

	scanner := bufio.NewScanner(f)  

	state := &State{balances, account2nonce, f, Block{}, Hash{}, false, miningDifficulty, gen.ForkBCIP1, map[uint64]int64{}}

	// set file position 
	filePos := int64(0) 

	for scanner.Scan(){
		if err := scanner.Err(); err != nil{
			return nil, err 
		}

		blockFsJson := scanner.Bytes() 

		if len(blockFsJson) == 0{
			break 
		}

		var blockFs BlockFS 
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil{
			return nil, err 
		}

		err = applyBlock(blockFs.Value, state) 
		if err != nil{
			return nil, err 
		}

		// set search caches
		state.HashCache[blockFs.Key.Hex()] = filePos 
		state.HeightCache[blockFs.Value.Header.Number] = filePos 
		filePos += int64(len(blockFsJson)) + 1 

		state.latestBlock = blockFs.Value  
		state.latestBlockHash = blockFs.Key 
		state.hasGenesisBlock = true 
	}

	return state, nil 
}

func (s *State) AddBlocks(blocks []Block) error{
	for _, b := range blocks{
		_, err := s.AddBlock(b) 
		if err != nil{
			return err
		}
	}
	return nil 
}

func (s *State) AddBlock(b Block) (Hash, error){
	pendingState := s.Copy()  

	err := applyBlock(b, &pendingState) 
	if err != nil{
		return Hash{}, err
	}

	blockHash, err := b.Hash() 
	if err != nil{
		return Hash{}, err 
	}

	blockFs := BlockFS{blockHash, b} 

	blockFsJson, err := json.Marshal(blockFs) 

	if err != nil{
		return Hash{}, err 
	}

	fmt.Printf("\nPersisting new block to disk:\n") 
	fmt.Printf("\t%s\n", blockFsJson)

	// Get file pos for cache 
	fs, _ := s.dbFile.Stat() 
	filePos := fs.Size() + 1 

	_, err = s.dbFile.Write(append(blockFsJson, '\n'))
	if err != nil{
		return Hash{}, err 
	}

	// set search caches 
	s.HashCache[blockFs.Key.Hex()] = filePos 
	s.HeightCache[blockFs.Value.Header.Number] = filePos 


	s.Balances = pendingState.Balances 
	s.Account2Nonce = pendingState.Account2Nonce 
	s.latestBlockHash = blockHash 
	s.latestBlock = b 
	s.hasGenesisBlock = true 
	s.miningDifficulty = pendingState.miningDifficulty

	return blockHash, nil 
}

func (s *State) NextBlockNumber() uint64{
	if !s.hasGenesisBlock{
		return uint64(0)
	}

	return s.LatestBlock().Header.Number + 1 


}

func (s *State) LatestBlock() Block{
	return s.latestBlock 
}

func (s *State) LatestBlockHash() Hash{
	return s.latestBlockHash
}

func (s *State) GetNextAccountNonce(account common.Address) uint{
	return s.Account2Nonce[account] + 1 
}

func (s *State) ChangeMiningDifficulty(newDifficulty uint){
	s.miningDifficulty = newDifficulty
}

func (s *State) IsBCIP1Fork() bool{
	return s.NextBlockNumber() >= s.forkBCL1
}

func (s *State) Copy() State{
	c := State{} 
	c.hasGenesisBlock = s.hasGenesisBlock 
	c.latestBlock = s.latestBlock 
	c.latestBlockHash = s.latestBlockHash 
	c.Balances = make(map[common.Address]uint) 
	c.Account2Nonce = make(map[common.Address]uint) 
	c.miningDifficulty = s.miningDifficulty 
	c.forkBCL1 = s.forkBCL1 


	for acc, balance := range s.Balances {
		c.Balances[acc] = balance 
	}

	for acc, nonce := range s.Account2Nonce{
		c.Account2Nonce[acc] = nonce 
	}

	return c 
}


func (s *State) Close() error{
	return s.dbFile.Close()
}

// applyBlock verifies if a block can be added to the blockchain 
func applyBlock(b Block, s *State) error{
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1 

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber{
		return fmt.Errorf("next expected block must be '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash){
		return fmt.Errorf("block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	hash, err := b.Hash() 

	if err != nil{
		return err 
	}

	if !IsBlockHashValid(hash, s.miningDifficulty){
		return fmt.Errorf("invalid block hash %x", hash)
	}


	err = applyTXs(b.TXs, s) 
	if err != nil{
		return err 
	}

	s.Balances[b.Header.Miner] += BlockReward 
	if s.IsBCIP1Fork(){
		s.Balances[b.Header.Miner] += b.BlockReward()
	} else{
		s.Balances[b.Header.Miner] += uint(len(b.TXs)) * TxFee
	}

	return nil 
}


func applyTXs(txs []SignedTx, s *State) error{
	//TODO: Implement tomorrow 
	return nil 
}