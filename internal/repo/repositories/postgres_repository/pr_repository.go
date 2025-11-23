package postgresrepository

import (
	"context"
	"strings"

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
	if result.Error != nil {
		le := strings.ToLower(result.Error.Error())
		if strings.Contains(le, "duplicate") || strings.Contains(le, "unique") || strings.Contains(le, "violates unique") {
			return ErrPRExists
		}
	}
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(pr).Error; err != nil {
			return err
		}

		if err := tx.Model(pr).Association("AssignedReviewers").Replace(pr.AssignedReviewers); err != nil {
			return err
		}

		return nil
	})
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

func (r *PReqRepository) ListOpenPullRequestsByReviewerIDs(ctx context.Context, reviewerIDs []string) ([]*models.PullRequest, error) {
	if len(reviewerIDs) == 0 {
		return nil, nil
	}
	var prs []*models.PullRequest
	result := r.db.WithContext(ctx).
		Preload("Author").
		Preload("AssignedReviewers").
		Joins("JOIN pull_request_reviewers ON pull_requests.id = pull_request_reviewers.pull_request_id").
		Where("pull_requests.status = ? AND pull_request_reviewers.user_id IN ?", "OPEN", reviewerIDs).
		Find(&prs)
	if result.Error != nil {
		return nil, result.Error
	}
	return prs, nil
}

func (r *PReqRepository) CountAssignmentsPerUser(ctx context.Context) (map[string]int64, error) {
	type row struct {
		UserCustomID string
		Cnt          int64
	}
	var rows []row
	q := r.db.WithContext(ctx).
		Table("pull_requests").
		Select("users.user_custom_id as user_custom_id, COUNT(*) as cnt").
		Joins("JOIN pull_request_reviewers ON pull_requests.id = pull_request_reviewers.pull_request_id").
		Joins("JOIN users ON pull_request_reviewers.user_id = users.id").
		Where("pull_requests.status = ?", "OPEN").
		Group("users.user_custom_id")

	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	res := make(map[string]int64, len(rows))
	for _, r := range rows {
		res[r.UserCustomID] = r.Cnt
	}
	return res, nil
}

func (r *PReqRepository) CountAssignmentsPerPR(ctx context.Context) (map[string]int64, error) {
	type row struct {
		PRCustomID string
		Cnt        int64
	}
	var rows []row

	q := r.db.WithContext(ctx).
		Table("pull_requests").
		Select("pull_requests.pull_request_custom_id as pr_custom_id, COUNT(pull_request_reviewers.user_id) as cnt").
		Joins("LEFT JOIN pull_request_reviewers ON pull_requests.id = pull_request_reviewers.pull_request_id").
		Group("pull_requests.pull_request_custom_id")

	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	res := make(map[string]int64, len(rows))
	for _, r := range rows {
		res[r.PRCustomID] = r.Cnt
	}
	return res, nil
}
