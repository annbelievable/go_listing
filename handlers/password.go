package handlers

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(password string) (string, error) {
	pwd := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return string(hash), nil
}

func ComparePasswords(hashedPassword string, password string) bool {
	byteHash := []byte(hashedPassword)
	pwd := []byte(password)
	err := bcrypt.CompareHashAndPassword(byteHash, pwd)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
