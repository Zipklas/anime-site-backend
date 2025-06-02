package comment

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(comment *Comment) error
	GetByAnimeID(animeID string, userID uuid.UUID) ([]CommentWithUser, error)
	Delete(commentID uuid.UUID, userID uuid.UUID) error
	Update(comment *Comment) error
	GetUserVote(commentID uuid.UUID, userID uuid.UUID) (*bool, error)
	GetVotes(commentID uuid.UUID) (upvotes, downvotes int, err error)
	AddVote(commentID uuid.UUID, userID uuid.UUID, isUpvote bool) error
	RemoveVote(commentID uuid.UUID, userID uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(comment *Comment) error {
	return r.db.Create(comment).Error
}

func (r *repository) AddVote(commentID uuid.UUID, userID uuid.UUID, isUpvote bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем предыдущий голос если был
		if err := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&CommentVote{}).Error; err != nil {
			return err
		}

		return tx.Create(&CommentVote{
			CommentID: commentID,
			UserID:    userID,
			IsUpvote:  isUpvote,
		}).Error
	})
}

func (r *repository) RemoveVote(commentID uuid.UUID, userID uuid.UUID) error {
	return r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&CommentVote{}).Error
}

func (r *repository) GetVotes(commentID uuid.UUID) (upvotes, downvotes int, err error) {
	var result struct {
		Upvotes   int
		Downvotes int
	}

	err = r.db.Model(&CommentVote{}).
		Select("SUM(CASE WHEN is_upvote = true THEN 1 ELSE 0 END) as upvotes, "+
			"SUM(CASE WHEN is_upvote = false THEN 1 ELSE 0 END) as downvotes").
		Where("comment_id = ?", commentID).
		Scan(&result).Error

	return result.Upvotes, result.Downvotes, err
}

func (r *repository) GetUserVote(commentID uuid.UUID, userID uuid.UUID) (*bool, error) {
	var vote CommentVote
	err := r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&vote).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vote.IsUpvote, nil
}

func (r *repository) GetByAnimeID(animeID string, userID uuid.UUID) ([]CommentWithUser, error) {
	var comments []CommentWithUser

	// Базовый запрос для комментариев
	baseQuery := r.db.Table("comments").
		Select("comments.*, users.email as user_email").
		Joins("left join users on comments.user_id = users.id").
		Where("comments.anime_id = ?", animeID).
		Order("comments.created_at desc")

	// Получаем комментарии
	if err := baseQuery.Scan(&comments).Error; err != nil {
		return nil, err
	}

	for i := range comments {
		up, down, err := r.GetVotes(comments[i].ID)
		if err != nil {
			return nil, err
		}

		comments[i].Upvotes = up
		comments[i].Downvotes = down

		if userID != uuid.Nil {
			vote, err := r.GetUserVote(comments[i].ID, userID)
			if err != nil {
				return nil, err
			}
			comments[i].UserVote = vote
		}
	}

	return comments, nil
}
func (r *repository) Delete(commentID uuid.UUID, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", commentID, userID).Delete(&Comment{}).Error
}

func (r *repository) Update(comment *Comment) error {
	return r.db.Model(comment).Where("id = ? AND user_id = ?", comment.ID, comment.UserID).Updates(comment).Error
}
