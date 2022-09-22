package file

import (
	"os"
	"os/user"
	"path"
	"strings"
)

// Expands a filepath
func ExpandPath(p string) string{
	if i := strings.Index(p, ":"); i > 0{
		return p 
	}
	if i := strings.Index(p, "@"); i > 0{
		return p 
	} 

	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\"){
		if home := 	homeDir(); home != ""{
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func RemoveDir(path string) error{
	return os.RemoveAll(path)
}

func homeDir() string{
	if home := os.Getenv("HOME"); home != ""{
		return home
	}
	if us, err := user.Current(); err == nil{
		return us.HomeDir
	}

	return ""
}