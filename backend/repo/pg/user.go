package pg

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	v1 "github.com/chaitin/panda-wiki/api/user/v1"
	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/pg"
)

type UserRepository struct {
	db     *pg.DB
	logger *log.Logger
}

func NewUserRepository(db *pg.DB, logger *log.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.WithModule("repo.pg.user"),
	}
}

func (r *UserRepository) UpsertDefaultUser(ctx context.Context, user *domain.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First try to find existing user
		var existingUser domain.User
		err := tx.Where("account = ?", user.Account).First(&existingUser).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// User doesn't exist, create new user
			if err := tx.Create(user).Error; err != nil {
				return err
			}
			return nil
		}
		// User exists, update password
		return tx.Model(&existingUser).Update("password", user.Password).Error
	})
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User, edition consts.LicenseEdition) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		limit, limited, message := licenseUserLimit(edition, user.Role)
		if limited {
			var count int64
			query := tx.Model(&domain.User{}).Where("role = ?", user.Role)
			if err := query.Count(&count).Error; err != nil {
				return err
			}
			if count >= limit {
				return errors.New(message)
			}
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
}

func licenseUserLimit(edition consts.LicenseEdition, role consts.UserRole) (limit int64, limited bool, message string) {
	switch role {
	case consts.UserRoleAdmin:
		switch edition {
		case consts.LicenseEditionFree:
			return 1, true, "free edition only allows 1 admin"
		case consts.LicenseEditionContributor:
			return 3, true, "contributor edition only allows 3 admins"
		}
	default:
		// no limit for non-admin users in current editions
	}
	return 0, false, ""
}

func (r *UserRepository) VerifyUser(ctx context.Context, account string, password string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("account = ?", account).First(&user).Error
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return &user, nil
}

func (r *UserRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]v1.UserListItemResp, error) {
	var users []v1.UserListItemResp
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) UpdateUserPassword(ctx context.Context, userID string, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("password", string(hashedPassword)).Error
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Delete(&domain.User{}).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Model(&domain.KBUsers{}).Where("user_id = ?", userID).Delete(&domain.KBUsers{}).Error; err != nil {
		return err
	}
	return nil
}
