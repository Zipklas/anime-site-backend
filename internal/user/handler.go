package user

import (
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Register(c echo.Context) error {
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.service.Register(req.Nickname, req.Email, req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "registered"})
}

func (h *Handler) Login(c echo.Context) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	log.Println(token)
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func (h *Handler) Profile(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	user, err := h.service.GetProfile(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"user_id":            user.ID,
		"email":              user.Email,
		"nickname":           user.Nickname,
		"avatar":             user.Avatar,
		"watched_anime_ids":  user.WatchedAnimeIDs,
		"favorite_anime_ids": user.FavoriteAnimeIDs,
	})
}

func (h *Handler) AddWatched(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	animeID := c.Param("anime_id")
	if animeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "anime_id is required")
	}

	if err := h.service.AddWatched(userID, animeID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":   "added to watched",
		"anime_id": animeID,
	})
}

func (h *Handler) AddFavorite(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	animeID := c.Param("anime_id")
	if animeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "anime_id is required")
	}

	if err := h.service.AddFavorite(userID, animeID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":   "added to favorites",
		"anime_id": animeID,
	})
}
func (h *Handler) GetWatchedAnime(c echo.Context) error {
	// Получаем userID из JWT токена
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	// Получаем список аниме с деталями
	animeList, err := h.service.GetWatchedAnimeDetails(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, animeList)
}
func (h *Handler) GetFavouriteAnime(c echo.Context) error {
	// Получаем userID из JWT токена
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	// Получаем список аниме с деталями
	animeList, err := h.service.GetFavouriteAnimeDetails(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, animeList)
}
func (h *Handler) UpdateNickname(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.service.UpdateNickname(userID, req.Nickname); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "nickname updated"})
}

func (h *Handler) UploadAvatar(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "user_id missing")
	}

	// Получаем файл из запроса
	file, err := c.FormFile("avatar")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "avatar file is required"})
	}

	// Проверяем размер файла (например, не более 2MB)
	if file.Size > 2<<20 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "file too large, max 2MB"})
	}

	// Открываем файл для проверки
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to open file"})
	}
	defer src.Close()

	// Декодируем изображение для проверки размеров
	img, _, err := image.Decode(src)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid image file"})
	}

	// Проверяем размеры (64x64)
	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "image must be 64x64 pixels"})
	}

	// Генерируем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	newFilename := uuid.New().String() + ext
	avatarPath := filepath.Join("uploads", "avatars", newFilename)

	// Создаем директории, если их нет
	if err := os.MkdirAll(filepath.Dir(avatarPath), 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create upload directory"})
	}

	// Сохраняем файл
	dst, err := os.Create(avatarPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to save file"})
	}
	defer dst.Close()

	// Сбрасываем позицию чтения файла
	if _, err := src.Seek(0, 0); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to reset file position"})
	}

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to save file"})
	}

	// Обновляем путь к аватару в базе данных
	if err := h.service.UpdateAvatar(userID, avatarPath); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to update avatar"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "avatar uploaded",
		"path":   avatarPath,
	})
}
