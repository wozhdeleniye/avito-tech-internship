package postgresrepository

import (
	"context"

	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	"gorm.io/gorm"
)

type PReqRepository struct {
	db *gorm.DB
}

func NewPReqRepository(db *gorm.DB) *PReqRepository {
	return &PReqRepository{db: db}
}

func (r *PReqRepository) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	result := r.db.WithContext(ctx).Create(pr)
	return result.Error
}

func (r *PReqRepository) GetPullRequestByID(ctx context.Context, id string) (*models.PullRequest, error) {
	var pr models.PullRequest
	result := r.db.WithContext(ctx).Preload("Author").Preload("AssignedReviewers").Where("pull_request_custom_id = ?", id).First(&pr)
	if result.Error != nil {
		return nil, result.Error
	}
	return &pr, nil
}

func (r *PReqRepository) UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	result := r.db.WithContext(ctx).Save(pr)
	return result.Error
}

func (r *PReqRepository) ListPullRequests(ctx context.Context) ([]*models.PullRequest, error) {
	var prs []*models.PullRequest

	result := r.db.WithContext(ctx).Preload("Author").Preload("AssignedReviewers").Find(&prs)
	if result.Error != nil {
		return nil, result.Error
	}
	return prs, nil
}

func (r *PReqRepository) ListPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	var prs []*models.PullRequest
	result := r.db.WithContext(ctx).Preload("Author").Preload("AssignedReviewers").Joins("JOIN pull_request_reviewers ON pull_requests.id = pull_request_reviewers.pull_request_id").Where("pull_request_reviewers.user_id = ?", reviewerID).Find(&prs)
	if result.Error != nil {
		return nil, result.Error
	}
	return prs, nil
}

func (r *PReqRepository) ListPullRequestsByReviewerCustomID(ctx context.Context, userCustomID string) ([]*models.PullRequest, error) {
	var prs []*models.PullRequest
	result := r.db.WithContext(ctx).
		Preload("Author").
		Preload("AssignedReviewers").
		Joins("JOIN pull_request_reviewers ON pull_requests.id = pull_request_reviewers.pull_request_id").
		Joins("JOIN users ON pull_request_reviewers.user_id = users.id").
		Where("users.user_custom_id = ?", userCustomID).
		Find(&prs)
	if result.Error != nil {
		return nil, result.Error
	}
	return prs, nil
}

func (r *PReqRepository) AssignReviewer(ctx context.Context, prID string, reviewer *models.User) error {
	var pr models.PullRequest
	result := r.db.WithContext(ctx).Preload("AssignedReviewers").Where("pull_request_custom_id = ?", prID).First(&pr)
	if result.Error != nil {
		return result.Error
	}
	pr.AssignedReviewers = append(pr.AssignedReviewers, reviewer)
	return r.db.WithContext(ctx).Save(&pr).Error
}
