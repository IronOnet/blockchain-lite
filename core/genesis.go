package core 

import (
	"encoding/json" 
	"ioutil" 
	
	"github.com/ethereum/go-ethereum/common"
)

var genesisJson = `{
		"genesis_time" : "2022-09-23T00:00:00.000000000Z", 
		"chain_id": "blockchain-lite-network", 
		"symbol": "BCL", 
		"balances": {
			"0x09eE50f2F37FcBA1845dE6FE5C762E83E65E755c": 1000000
		}, 
		"fork_bcip_1": 35
	}`

type Genesis struct{
	Balances map[common.Address]uint `json:"balances"`
	Symbol string 					 `json:"symbol"`


	ForkBCIP1 uint64 				 `json:"fork_bcip_1"`
}


func loadGenesis(path string) (Genesis, error){
	content, err := ioutil.ReadFile(path) 
	if err != nil{
		return Genesis{}, err 
	}

	var loadGenesis Genesis 
	err = json.Unmarshal(content, &loadGenesis) 
	if err != nil{
		return Genesis{}, err 
	}

	return loadGenesis, nil 
} 


func writeGenesisToDisk(path string, genesis[] byte) error{
	return ioutil.WriteFile(path, genesis, 0644)
}

