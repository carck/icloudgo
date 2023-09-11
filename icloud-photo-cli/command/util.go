package command

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"golang.org/x/term"
	"os"
	"syscall"
)

const (
	hashSize = 16 * 1024
)

// Hash returns the SHA1 hash of a file as string.
func Hash(fileName string) (string, error) {
	if bytes, err := readHashBytes(fileName); err != nil {
		return "", err
	} else {
		hash := sha1.New()
		if _, hErr := hash.Write(bytes); hErr != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	}
}

func readHashBytes(filePath string) ([]byte, error) {
	if fi, err := os.Stat(filePath); err != nil {
		return nil, err
	} else if fi.Size() <= hashSize {
		return os.ReadFile(filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	firstBytes := make([]byte, hashSize/2)
	if _, e := file.ReadAt(firstBytes, 0); e != nil {
		return nil, fmt.Errorf("couldn't read first few bytes: %+v", e)
	}
	middleBytes := make([]byte, hashSize/4)
	fileInfo, _ := file.Stat()
	if _, e := file.ReadAt(middleBytes, fileInfo.Size()/2); e != nil {
		return nil, fmt.Errorf("couldn't read middle bytes: %+v", e)
	}
	lastBytes := make([]byte, hashSize/4)
	if _, e := file.ReadAt(lastBytes, fileInfo.Size()-hashSize/4); e != nil {
		return nil, fmt.Errorf("couldn't read end bytes: %+v", e)
	}
	bytes := append(append(firstBytes, middleBytes...), lastBytes...)
	return bytes, nil
}

func getTextInput(tip, defaultValue string, secure bool) func(string) (string, error) {
	return func(string2 string) (string, error) {
		if defaultValue != "" {
			return defaultValue, nil
		}
		fmt.Println("Please input", tip)
		if secure {
			password, err := term.ReadPassword(int(syscall.Stdin))
			return string(password), err
		}
		var s string
		_, err := fmt.Scanln(&s)
		return s, err
	}
}

func mkdirAll(path string) error {
	if f, _ := os.Stat(path); f == nil {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
