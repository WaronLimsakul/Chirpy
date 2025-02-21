package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	cost := 10 // iteration of encryption
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// return nil if password is correct
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
