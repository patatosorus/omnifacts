package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword génère un hash bcrypt à partir d'un mot de passe en clair
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword vérifie qu'un mot de passe correspond à son hash
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
