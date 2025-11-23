package models

import (
	"time"

	"github.com/google/uuid"
	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	"gorm.io/gorm"
)

type PullRequest struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PullRequestCustomID string    `gorm:"uniqueIndex;not null" json:"pull_request_custom_id"`
	PullRequestName     string    `gorm:"not null" json:"pull_request_name"`
	AuthorID            uuid.UUID `gorm:"type:uuid;not null" json:"author_id"`
	Author              User      `gorm:"foreignKey:AuthorID" json:"-"`
	Status              string    `gorm:"type:varchar(10);not null;default:'OPEN'" json:"status"`
	AssignedReviewers   []*User   `gorm:"many2many:pull_request_reviewers;" json:"assigned_reviewers"`
	MergedAt            *int64    `json:"mergedAt,omitempty"`
	CreatedAt           int64     `gorm:"autoCreateTime" json:"createdAt"`
}

func (p *PullRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.CreatedAt = time.Now().Unix()
	return nil
}

type PullRequestReassign struct {
	PullRequest   openapi.PullRequest `json:"pr"`
	NewReviewerID string              `json:"replaced_by"`
}

type PullRequestSearch struct {
	PullRequest []*openapi.PullRequestShort `json:"pull_requests"`
	Author      string                      `json:"user_id"`
}
