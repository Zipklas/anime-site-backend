package shikimori

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SearchAnime(c echo.Context) error {

	search := c.QueryParam("search")
	if search == "" {
		search = "bakemono"
	}

	log.Printf("Поиск аниме по запросу: %s", search)

	animes, err := h.service.SearchAnime(c.Request().Context(), search, 10)
	if err != nil {

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(animes) == 0 {
		log.Println("Не найдено аниме по запросу.")
	}

	// Возвращаем найденные аниме
	return c.JSON(http.StatusOK, animes)
}

// GetTopAnime - обработчик для получения топовых аниме по рейтингу
func (h *Handler) GetTopAnime(c echo.Context) error {

	limitStr := c.QueryParam("limit")
	pageStr := c.QueryParam("page")

	limit := 30
	page := 1

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	log.Printf("Запрос топ аниме: limit=%d, page=%d", limit, page)

	animes, err := h.service.GetTopAnime(c.Request().Context(), limit, page)
	if err != nil {
		log.Printf("Ошибка при получении топ-аниме: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось получить топ-аниме"})
	}

	return c.JSON(http.StatusOK, animes)
}
func (h *Handler) GetAnimeByID(c echo.Context) error {
	animeID := c.Param("id")
	if animeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "anime ID is required"})
	}

	anime, err := h.service.GetAnimeByID(c.Request().Context(), animeID)
	if err != nil {
		log.Printf("Ошибка при получении аниме: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось получить информацию об аниме"})
	}

	return c.JSON(http.StatusOK, anime)
}
func (h *Handler) GetNewReleases(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 10 // значение по умолчанию

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	animes, err := h.service.GetNewReleases(c.Request().Context(), limit)
	if err != nil {
		log.Printf("Ошибка при получении новинок: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Не удалось получить список новинок",
		})
	}

	return c.JSON(http.StatusOK, animes)
}
