package postgresrepository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	return result.Error
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ? AND is_active = ?", email, true).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Save(user)
	return result.Error
}

func (r *UserRepository) UserExists(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ? AND is_active = ?", email, true).Count(&count)
	return count > 0, result.Error
}

func (r *UserRepository) GetUserByCustomID(ctx context.Context, customID string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("user_custom_id = ? AND is_active = ?", customID, true).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) UserExistsByCustomID(ctx context.Context, customID string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("user_custom_id = ? AND is_active = ?", customID, true).Count(&count)
	return count > 0, result.Error
}
