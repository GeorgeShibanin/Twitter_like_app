package generator

import (
	"encoding/base64"
	"math/rand"
)

var Alphabet = []byte("qwertyuiopasdfghjklzxcvbnm1234567890")

func GetUserRandomId(size int) string {
	idBytes := make([]byte, size)
	for i := 0; i < len(idBytes); i++ {
		idBytes[i] = Alphabet[rand.Intn(len(Alphabet))]
	}
	return string(idBytes)
}

func GenerateBase64ID(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	return encoded, nil
}
