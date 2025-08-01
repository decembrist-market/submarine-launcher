package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

func checkHash(archivePath string) error {
	remoteHash, err := downloadRemoteHash()
	if err != nil {
		return err
	}
	localHash, err := calcFileSHA256(archivePath)
	if err != nil {
		return err
	}
	if remoteHash != localHash {
		return fmt.Errorf("хеш архива не совпадает")
	}
	return nil
}

func downloadRemoteHash() (string, error) {
	hashUrl := GetHashURL()
	resp, err := http.Get(hashUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("сервер вернул статус %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(trimHash(data)), nil
}

func trimHash(data []byte) []byte {
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r' || data[len(data)-1] == ' ') {
		data = data[:len(data)-1]
	}
	return data
}

func calcFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sha := sha256.New()
	if _, err := io.Copy(sha, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(sha.Sum(nil)), nil
}
