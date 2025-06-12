package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Zipklas/anime-site-backend/internal/shikimori"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(nickname, email, password string) error
	Login(email, password string) (string, error)
	UpdateNickname(userID string, nickname string) error
	UpdateAvatar(userID string, avatarPath string) error
	GetProfile(userID string) (*User, error)
	AddWatched(userID, animeID string) error
	AddFavorite(userID, animeID string) error
	GetAnimeLists(userID string) (watched []shikimori.Anime, favorites []shikimori.Anime, err error)
	GetWatchedAnimeDetails(userID string) ([]shikimori.Anime, error)
	GetFavouriteAnimeDetails(userID string) ([]shikimori.Anime, error)
}

type service struct {
	repo             Repository
	shikimoriService *shikimori.Service
}

func NewService(repo Repository, shikimoriService *shikimori.Service) Service {
	return &service{
		repo:             repo,
		shikimoriService: shikimoriService,
	}
}

func (s *service) GetAnimeLists(userID string) ([]shikimori.Anime, []shikimori.Anime, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, nil, err
	}

	var watched []shikimori.Anime
	for _, id := range user.WatchedAnimeIDs {
		anime, err := s.shikimoriService.GetAnimeByID(context.Background(), id)
		if err == nil {
			watched = append(watched, *anime)
		}
	}

	var favorites []shikimori.Anime
	for _, id := range user.FavoriteAnimeIDs {
		anime, err := s.shikimoriService.GetAnimeByID(context.Background(), id)
		if err == nil {
			favorites = append(favorites, *anime)
		}
	}

	return watched, favorites, nil
}

func (s *service) Register(nickname, email, password string) error {
	// Генерация UUID для нового пользователя
	id := uuid.New()
	if len(nickname) > 32 {
		return errors.New("nickname too long, max 32 characters")
	}
	if nickname == "" {
		return errors.New("nickname cannot be empty")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &User{
		ID:       id,
		Email:    email,
		Nickname: nickname,
		Password: string(hash),
	}
	return s.repo.Create(user)
}

func (s *service) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // токен действителен 72 часа
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT secret is not set")
	}
	log.Println("JWT_SECRET:", os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	log.Println("Generated Token:", tokenString)

	return tokenString, nil
}
func (s *service) GetProfile(userID string) (*User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
func (s *service) AddWatched(userID, animeID string) error {
	return s.repo.UpdateWatched(userID, animeID)
}

func (s *service) AddFavorite(userID, animeID string) error {
	return s.repo.UpdateFavorites(userID, animeID)
}
func (s *service) GetWatchedAnimeDetails(userID string) ([]shikimori.Anime, error) {
	// Получаем пользователя с его списком просмотренных аниме
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Если список пуст, возвращаем пустой массив
	if len(user.WatchedAnimeIDs) == 0 {
		return []shikimori.Anime{}, nil
	}

	// Получаем информацию об аниме из Shikimori API
	var animeList []shikimori.Anime
	for _, animeID := range user.WatchedAnimeIDs {
		anime, err := s.shikimoriService.GetAnimeByID(context.Background(), animeID)
		if err != nil {
			log.Printf("Failed to get anime %s: %v", animeID, err)
			continue // Пропускаем если не удалось получить
		}
		animeList = append(animeList, *anime)
	}

	return animeList, nil
}
func (s *service) GetFavouriteAnimeDetails(userID string) ([]shikimori.Anime, error) {
	// Получаем пользователя с его списком просмотренных аниме
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Если список пуст, возвращаем пустой массив
	if len(user.FavoriteAnimeIDs) == 0 {
		return []shikimori.Anime{}, nil
	}

	// Получаем информацию об аниме из Shikimori API
	var animeList []shikimori.Anime
	for _, animeID := range user.FavoriteAnimeIDs {
		anime, err := s.shikimoriService.GetAnimeByID(context.Background(), animeID)
		if err != nil {
			log.Printf("Failed to get anime %s: %v", animeID, err)
			continue // Пропускаем если не удалось получить
		}
		animeList = append(animeList, *anime)
	}

	return animeList, nil
}
func (s *service) UpdateNickname(userID string, nickname string) error {
	// Можно добавить валидацию никнейма
	if len(nickname) > 32 {
		return errors.New("nickname too long, max 32 characters")
	}
	if nickname == "" {
		return errors.New("nickname cannot be empty")
	}
	return s.repo.UpdateNickname(userID, nickname)
}

func (s *service) UpdateAvatar(userID string, avatarPath string) error {
	return s.repo.UpdateAvatar(userID, avatarPath)
}
