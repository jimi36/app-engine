package utils

import (
	"io/ioutil"
	"os"
)

func IsExistedPath(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	st, err := os.Stat(path)
	if err != nil || st == nil || !st.IsDir() {
		return false
	}
	return true
}

func CreateFolder(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func RemoveFolder(path string) error {
	return os.RemoveAll(path)
}

func CreateFile(filePath string, data []byte) error {
	if IsExistedPath(filePath) {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(filePath, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}
