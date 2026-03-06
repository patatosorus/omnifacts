package db

import (
	"omnifacts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionDB interface {
	Create(perm *models.Permission) error
	FindByUserAndRepo(userID, repoID uuid.UUID) (*models.Permission, error)
	ListByUser(userID uuid.UUID) ([]models.Permission, error)
	ListByRepo(repoID uuid.UUID) ([]models.Permission, error)
	Update(perm *models.Permission) error
	Delete(id uuid.UUID) error
	DeleteByUserAndRepo(userID, repoID uuid.UUID) error
}

type permissionDB struct {
	db *gorm.DB
}

func NewPermissionDB(db *gorm.DB) PermissionDB {
	return &permissionDB{db: db}
}

func (r *permissionDB) Create(perm *models.Permission) error {
	return r.db.Create(perm).Error
}

func (r *permissionDB) FindByUserAndRepo(userID, repoID uuid.UUID) (*models.Permission, error) {
	var perm models.Permission
	err := r.db.First(&perm, "user_id = ? AND repository_id = ?", userID, repoID).Error
	return &perm, err
}

func (r *permissionDB) ListByUser(userID uuid.UUID) ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.Preload("Repository").Where("user_id = ?", userID).Find(&perms).Error
	return perms, err
}

func (r *permissionDB) ListByRepo(repoID uuid.UUID) ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.Preload("User").Where("repository_id = ?", repoID).Find(&perms).Error
	return perms, err
}

func (r *permissionDB) Update(perm *models.Permission) error {
	return r.db.Save(perm).Error
}

func (r *permissionDB) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

func (r *permissionDB) DeleteByUserAndRepo(userID, repoID uuid.UUID) error {
	return r.db.Where("user_id = ? AND repository_id = ?", userID, repoID).Delete(&models.Permission{}).Error
}
