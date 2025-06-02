package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CommentModerationResult struct {
	IsApproved    bool               `json:"is_approved"`
	ToxicityScore float64            `json:"toxicity_score"`
	Details       map[string]float64 `json:"details"` // Изменяем тип для удобства работы
}

type Service interface {
	CreateComment(ctx context.Context, animeID, content string, userID uuid.UUID, parentID *uuid.UUID) (*Comment, error)
	GetComments(ctx context.Context, animeID string, userID uuid.UUID) ([]CommentWithUser, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error
	UpdateComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, content string) error
	VoteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, isUpvote bool) error
	RemoveVote(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error
	moderateComment(content string) (*CommentModerationResult, error)
}

type service struct {
	repo          Repository
	moderationURL string
	httpClient    *http.Client
}

func NewService(repo Repository) Service {
	return &service{
		repo:          repo,
		moderationURL: os.Getenv("MODERATION_SERVICE_URL"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *service) moderateComment(content string) (*CommentModerationResult, error) {
	if s.moderationURL == "" {
		return &CommentModerationResult{IsApproved: true}, nil
	}

	requestBody, err := json.Marshal(map[string]string{"text": content})
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Post(
		s.moderationURL+"/moderate",
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CommentModerationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *service) CreateComment(ctx context.Context, animeID, content string, userID uuid.UUID, parentID *uuid.UUID) (*Comment, error) {
	if content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	if len(content) > 1000 {
		return nil, errors.New("comment is too long")
	}

	moderation, err := s.moderateComment(content)
	if err != nil {
		return nil, errors.New("moderation service error")
	}

	if !moderation.IsApproved {

		var toxicLabels []string
		for label, score := range moderation.Details {
			if score > 0.5 && label != "non-toxic" { // Показываем только токсичные категории с высоким скором
				toxicLabels = append(toxicLabels, fmt.Sprintf("%s (%.0f%%)", label, score*100))
			}
		}

		errorMsg := fmt.Sprintf(
			"Ваш комментарий был отклонен системой модерации. "+
				"Общий уровень токсичности: %.0f%%. "+
				"Проблемные категории: %s. "+
				"Пожалуйста, переформулируйте ваш комментарий.",
			moderation.ToxicityScore*100,
			strings.Join(toxicLabels, ", "),
		)
		return nil, errors.New(errorMsg)
	}

	comment := &Comment{
		ID:         uuid.New(),
		AnimeID:    animeID,
		UserID:     userID,
		Content:    content,
		ParentID:   parentID,
		IsApproved: moderation.IsApproved,
	}

	if err := s.repo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *service) VoteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, isUpvote bool) error {

	return s.repo.AddVote(commentID, userID, isUpvote)
}

func (s *service) RemoveVote(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	return s.repo.RemoveVote(commentID, userID)
}

func (s *service) GetComments(ctx context.Context, animeID string, userID uuid.UUID) ([]CommentWithUser, error) {
	return s.repo.GetByAnimeID(animeID, userID)
}

func (s *service) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(commentID, userID)
}

func (s *service) UpdateComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, content string) error {
	if content == "" {
		return errors.New("comment content cannot be empty")
	}

	return s.repo.Update(&Comment{
		ID:      commentID,
		UserID:  userID,
		Content: content,
	})
}
