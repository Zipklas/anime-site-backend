package comment

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateComment(c echo.Context) error {
	animeID := c.Param("anime_id")
	if animeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "anime_id is required")
	}

	var req struct {
		Content  string     `json:"content"`
		ParentID *uuid.UUID `json:"parent_id,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, err := getUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	comment, err := h.service.CreateComment(c.Request().Context(), animeID, req.Content, userID, req.ParentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, comment)
}

func (h *Handler) VoteComment(c echo.Context) error {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment_id")
	}

	var req struct {
		IsUpvote bool `json:"is_upvote"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, err := getUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.service.VoteComment(c.Request().Context(), commentID, userID, req.IsUpvote); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RemoveVote(c echo.Context) error {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment_id")
	}

	userID, err := getUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.service.RemoveVote(c.Request().Context(), commentID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetComments(c echo.Context) error {
	animeID := c.Param("anime_id")
	if animeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "anime_id is required")
	}

	userID, _ := getUserIDFromToken(c) // Ошибка не критична - просто не будет user_vote

	comments, err := h.service.GetComments(c.Request().Context(), animeID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, comments)
}

func (h *Handler) DeleteComment(c echo.Context) error {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment_id")
	}

	userID, err := getUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.service.DeleteComment(c.Request().Context(), commentID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UpdateComment(c echo.Context) error {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment_id")
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, err := getUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.service.UpdateComment(c.Request().Context(), commentID, userID, req.Content); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func getUserIDFromToken(c echo.Context) (uuid.UUID, error) {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userIDStr := claims["user_id"].(string)
	return uuid.Parse(userIDStr)
}
