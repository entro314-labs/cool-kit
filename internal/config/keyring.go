package config

import (
	"github.com/99designs/keyring"
)

const (
	keyringService = "cool-kit"
)

// getKeyring returns a new keyring instance
func getKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: keyringService,
	})
}

// setToken saves a token to the keyring
func setToken(key, token string) error {
	kr, err := getKeyring()
	if err != nil {
		return err
	}

	return kr.Set(keyring.Item{
		Key:  key,
		Data: []byte(token),
	})
}

// getToken retrieves a token from the keyring
func getToken(key string) (string, error) {
	kr, err := getKeyring()
	if err != nil {
		return "", err
	}

	item, err := kr.Get(key)
	if err != nil {
		return "", err
	}

	return string(item.Data), nil
}

