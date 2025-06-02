package comment

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	Votes      []CommentVote `gorm:"foreignKey:CommentID"`
	ID         uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	AnimeID    string        `gorm:"index" json:"anime_id"` // Shikimori ID аниме
	UserID     uuid.UUID     `json:"user_id"`
	Content    string        `gorm:"type:text" json:"content"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	ParentID   *uuid.UUID    `gorm:"type:uuid;index" json:"parent_id,omitempty"` // Для ответов на комментарии
	IsApproved bool          `gorm:"default:true" json:"is_approved"`
}

type CommentVote struct {
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid" json:"-"`
	CommentID uuid.UUID `gorm:"primaryKey;type:uuid" json:"comment_id"`
	IsUpvote  bool      `json:"is_upvote"` // true - лайк, false - дизлайк
	Comment   Comment   `gorm:"foreignKey:CommentID"`
}

type CommentWithUser struct {
	Comment
	UserEmail string `json:"user_email"`
	Upvotes   int    `json:"upvotes"`
	Downvotes int    `json:"downvotes"`
	UserVote  *bool  `json:"user_vote"` // nil - нет голоса, true - лайк, false - дизлайк
}
