package shortcode

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
)

func GenerateShortURL() string {
	codeLength, err := strconv.Atoi(os.Getenv("CODE_LENGTH"))
	if err != nil {
		return ""
	}
	b := make([]byte, codeLength)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:codeLength]
}
