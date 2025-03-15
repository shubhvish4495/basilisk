package helper

import "os"

func CheckFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
