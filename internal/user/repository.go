package user

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Repository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(userID string) (*User, error)

	UpdateWatched(userID string, animeID string) error
	UpdateFavorites(userID string, animeID string) error

	UpdateNickname(userID string, nickname string) error
	UpdateAvatar(userID string, avatarPath string) error
}
type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}
func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}
func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(userID string) (*User, error) {
	var user User
	result := r.db.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	if user.WatchedAnimeIDs == nil {
		user.WatchedAnimeIDs = pq.StringArray{}
	}
	if user.FavoriteAnimeIDs == nil {
		user.FavoriteAnimeIDs = pq.StringArray{}
	}

	return &user, nil
}
func (r *repository) UpdateWatched(userID string, animeID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}

		// Проверяем, есть ли уже этот animeID в массиве
		for _, id := range user.WatchedAnimeIDs {
			if id == animeID {
				return nil
			}
		}

		// Обновляем массив
		user.WatchedAnimeIDs = append(user.WatchedAnimeIDs, animeID)
		return tx.Save(&user).Error
	})
}

func (r *repository) UpdateFavorites(userID string, animeID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}

		// Проверяем на существование
		for _, id := range user.FavoriteAnimeIDs {
			if id == animeID {
				return nil
			}
		}

		user.FavoriteAnimeIDs = append(user.FavoriteAnimeIDs, animeID)
		return tx.Save(&user).Error
	})
}
func (r *repository) UpdateNickname(userID string, nickname string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("nickname", nickname).Error
}

func (r *repository) UpdateAvatar(userID string, avatarPath string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("avatar", avatarPath).Error
}
