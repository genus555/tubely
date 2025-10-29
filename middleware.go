package main

import (
	"encoding/base64"
	"crypto/rand"
)

func MakeFileName() string {
	key := make([]byte, 32)
	rand.Read(key)
	file_name := base64.RawURLEncoding.EncodeToString(key)

	return file_name
}