package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/utils"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(username, email, password string) (*models.User, error)
	Login(username, password string) (string, *models.User, error)
	GenerateAPIKey(userID uuid.UUID, name string) (string, *models.APIKey, error)
	ListAPIKeys(userID uuid.UUID) ([]models.APIKey, error)
	DeleteAPIKey(userID, keyID uuid.UUID) error
}

type authService struct {
	userDB   db.UserDB
	apiKeyDB db.APIKeyDB
}

func NewAuthService(userDB db.UserDB, apiKeyDB db.APIKeyDB) AuthService {
	return &authService{userDB: userDB, apiKeyDB: apiKeyDB}
}

func (s *authService) Register(username, email, password string) (*models.User, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("impossible de hasher le mot de passe : %w", err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleUser,
		Active:       true,
	}

	if err := s.userDB.Create(user); err != nil {
		return nil, fmt.Errorf("impossible de créer l'utilisateur : %w", err)
	}

	return user, nil
}

func (s *authService) Login(username, password string) (string, *models.User, error) {
	user, err := s.userDB.FindByUsername(username)
	if err != nil {
		return "", nil, fmt.Errorf("identifiants invalides")
	}

	if !user.Active {
		return "", nil, fmt.Errorf("compte désactivé")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", nil, fmt.Errorf("identifiants invalides")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return "", nil, fmt.Errorf("impossible de générer le token : %w", err)
	}

	return token, user, nil
}

func (s *authService) GenerateAPIKey(userID uuid.UUID, name string) (string, *models.APIKey, error) {
	rawKey, err := generateRandomKey(32)
	if err != nil {
		return "", nil, fmt.Errorf("impossible de générer la clé : %w", err)
	}

	prefixedKey := "omni_" + rawKey
	keyHash := hashKey(prefixedKey)

	apiKey := &models.APIKey{
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: prefixedKey[:10],
		UserID:    userID,
	}

	if err := s.apiKeyDB.Create(apiKey); err != nil {
		return "", nil, fmt.Errorf("impossible de sauvegarder la clé API : %w", err)
	}

	return prefixedKey, apiKey, nil
}

func (s *authService) ListAPIKeys(userID uuid.UUID) ([]models.APIKey, error) {
	return s.apiKeyDB.ListByUser(userID)
}

func (s *authService) DeleteAPIKey(userID, keyID uuid.UUID) error {
	return s.apiKeyDB.Delete(keyID)
}

func generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)
}
