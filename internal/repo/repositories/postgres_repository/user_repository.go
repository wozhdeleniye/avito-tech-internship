package postgresrepository

import (
	"context"
	"errors"

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
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepository) GetUserByCustomId(ctx context.Context, cutstomId string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("user_custom_id = ?", cutstomId).First(&user) //доделать проверку с custom id

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

func (r *UserRepository) DeleteUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Delete(user)
	return result.Error
}

func (r *UserRepository) GetUserByCustomIDActive(ctx context.Context, customID string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("user_custom_id = ? AND is_active = ?", customID, true).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
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
