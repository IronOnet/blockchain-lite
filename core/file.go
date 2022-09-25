package core 

import (
	"io/ioutil" 
	"os" 
	"path/filepath"
)


func InitDirIfNotExist(dataDir string, genesis []byte) error{
	if fileExist(getGenesisJsonFilePath(dataDir)){
		return nil 
	}

	if err := os.MkdirAll(getDatabaseDirPath(dataDir), os.ModePerm); err != nil{
		return err 
	}

	if err:= writeGenesisToDisk(getGenesisJsonFilePath(dataDir), genesis); err!= nil{
		return err 
	}

	if err := writeEmptyBlockDbToDisk(getBlocksDBFilePath(dataDir)); err != nil{
		return err 
	}

	return nil 
}

func getDatabaseDirPath(dataDir string) string{
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonFilePath(dataDir string) string{
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDBFilePath(dataDir string) string{
	return filepath.Join(getDatabaseDirPath(dataDir), "block.db")
}

func fileExist(filePath string) bool{
	_, err := os.Stat(filePath) 
	if err != nil && os.IsNotExist(err){
		return false 
	}
	return true 
}

func dirExist(path string) (bool, error){
	_, err := os.Stat(path) 
	if err == nil{
		return true, nil 
	}

	if os.IsNotExist(err){
		return false, nil 
	}

	return true, err 
}

func writeEmptyBlockDbToDisk(path string) error{
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}