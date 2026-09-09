package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const LegacyHardcodedKey = "xxangLiJi2026_SecretKey_32Bytes!"

var (
	mu      sync.RWMutex
	primary []byte
	legacy  = []byte(LegacyHardcodedKey)
)

func SetPrimary(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("AES primary key is empty")
	}
	b := []byte(key)
	if n := len(b); n != 16 && n != 24 && n != 32 {
		return fmt.Errorf("AES primary key length %d is not 16/24/32", n)
	}
	if key == LegacyHardcodedKey {
		return errors.New("AES primary key must not be the hardcoded default")
	}
	mu.Lock()
	primary = append([]byte(nil), b...)
	mu.Unlock()
	return nil
}

func InitFromConfig() error {
	key := strings.TrimSpace(os.Getenv("PHONE_AES_KEY"))
	if key == "" {
		key = strings.TrimSpace(viper.GetString("encryptKey.data"))
	}
	return SetPrimary(key)
}

func Encrypt(plain string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", errors.New("plaintext is empty")
	}
	mu.RLock()
	key := append([]byte(nil), primary...)
	mu.RUnlock()
	if len(key) == 0 {
		return "", errors.New("AES primary key is not set")
	}
	return encryptWith(key, plain)
}

func Decrypt(ciphertext string) (string, error) {
	ciphertext = strings.TrimSpace(ciphertext)
	if ciphertext == "" {
		return "", errors.New("ciphertext is empty")
	}
	mu.RLock()
	key := append([]byte(nil), primary...)
	mu.RUnlock()
	if len(key) > 0 {
		if plain, err := decryptWith(key, ciphertext); err == nil {
			return plain, nil
		}
	}
	return decryptWith(legacy, ciphertext)
}

func GetDataSuffix(data string) string {
	data = strings.TrimSpace(data)
	if len(data) < 4 {
		return data
	}
	return data[len(data)-4:]
}

func GetDataSec(data string) string {
	data = strings.TrimSpace(data)
	if len(data) != 11 {
		return data
	}
	return data[:3] + "****" + data[7:]
}

func GetDataHash(data string) string {
	data = strings.TrimSpace(data)
	if data == "" {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func encryptWith(key []byte, plain string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

func decryptWith(key []byte, encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	plain, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func Hash(data string) string {
	data = strings.TrimSpace(data)
	if data == "" {
		return ""
	}
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func Suffix(data string, n int) string {
	data = strings.TrimSpace(data)
	if n <= 0 || len(data) <= n {
		return data
	}
	return data[len(data)-n:]
}
