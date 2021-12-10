package generator

import (
	"encoding/base64"
	"math/rand"
	"time"
)

var Alphabet = []byte("qwertyuiopasdfghjklzxcvbnm1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GetRandomId(size int) []byte {
	idBytes := make([]byte, size)
	for i := 0; i < len(idBytes); i++ {
		idBytes[i] = Alphabet[rand.Intn(len(Alphabet))]
	}
	return idBytes
}

func GenerateBase64ID(size int) (string, error) {
	b := make([]byte, size)
	rand.Seed(time.Now().UTC().UnixNano())
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	//prefixedID := "inMem" + encoded
	return encoded, nil
}
