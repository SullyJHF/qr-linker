package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

const (
	hashLength = 6
	maxRetries = 5
)

func GenerateShortHash() (string, error) {
	bytes := make([]byte, hashLength)
	
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	hash := base64.URLEncoding.EncodeToString(bytes)
	hash = strings.TrimRight(hash, "=")
	
	if len(hash) > hashLength {
		hash = hash[:hashLength]
	}
	
	return hash, nil
}

func GenerateUniqueHash(checkExists func(string) (bool, error)) (string, error) {
	for i := 0; i < maxRetries; i++ {
		hash, err := GenerateShortHash()
		if err != nil {
			return "", err
		}
		
		exists, err := checkExists(hash)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return hash, nil
		}
	}
	
	for length := hashLength + 1; length <= hashLength+4; length++ {
		bytes := make([]byte, length)
		_, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}
		
		hash := base64.URLEncoding.EncodeToString(bytes)
		hash = strings.TrimRight(hash, "=")
		if len(hash) > length {
			hash = hash[:length]
		}
		
		exists, err := checkExists(hash)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return hash, nil
		}
	}
	
	return "", nil
}