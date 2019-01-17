package utils

import (
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
)

func RIPEMD160(data []byte) []byte {
	hash := ripemd160.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func SHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}