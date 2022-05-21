package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
)

func Encrypt(data []byte, keyString string) ([]byte, error) {
	key, _ := hex.DecodeString(keyString)
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, errors.New("Error: Unable to encrypt the data.")
	}

	aesGCM, err := cipher.NewGCM(block)

	if err != nil {
		return nil, errors.New("Error: Unable to encrypt the data.")
	}

	nonce := make([]byte, aesGCM.NonceSize())
	cipherData := aesGCM.Seal(nonce, nonce, data, nil)

	return cipherData, nil
}

func Decrypt(encryptedData []byte, keyString string) ([]byte, error) {
	key, _ := hex.DecodeString(keyString)
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, errors.New("Error: Unable to decrypt the data. The key is might be incorrect.")
	}

	aesGCM, err := cipher.NewGCM(block)

	if err != nil {
		return nil, errors.New("Error: Unable to decrypt the data. The key is might be incorrect.")
	}

	nonceSize := aesGCM.NonceSize()
	nonce, cipherData := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plainData, err := aesGCM.Open(nil, nonce, cipherData, nil)

	if err != nil {
		return nil, errors.New("Error: Unable to decrypt the data. The key is might be incorrect.")
	}

	return plainData, nil
}

func LockFile(file *File, keyString string) error {
	if file.locked {
		return errors.New("Error: The file is already locked.")
	}

	file.key = keyString

	file.locked = true

	return nil
}

func LockFolder(folder *Folder, keyString string) error {
	if folder.locked {
		return errors.New("Error: The folder is already locked.")
	}

	folder.key = keyString

	folder.locked = true

	return nil
}

func UnlockFile(file *File, keyString string) error {
	if !file.locked {
		return errors.New("Error: The file is not locked.")
	}

	if file.key != keyString {
		return errors.New("Error: Unable to unlock the file. The key might be incorrect.")
	}

	file.locked = false

	return nil
}

func UnlockFolder(folder *Folder, keyString string) error {
	if !folder.locked {
		return errors.New("Error: The folder is not locked.")
	}

	if folder.key != keyString {
		return errors.New("Error: Unable to unlock the folder. The key might be incorrect.")
	}

	folder.locked = false

	return nil
}
