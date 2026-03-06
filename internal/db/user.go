package db

import (
	"omnifacts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDB interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	List() ([]models.User, error)
}

type userDB struct {
	db *gorm.DB
}

func NewUserDB(db *gorm.DB) UserDB {
	return &userDB{db: db}
}

func (r *userDB) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userDB) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	return &user, err
}

func (r *userDB) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "username = ?", username).Error
	return &user, err
}

func (r *userDB) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	return &user, err
}

func (r *userDB) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userDB) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userDB) List() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}
