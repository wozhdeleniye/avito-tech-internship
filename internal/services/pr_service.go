package services

import (
	"context"
	"errors"
	"time"

	"math/rand"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	serviceerrors "github.com/wozhdeleniye/avito-tech-internship/internal/pkg/errors"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	postgresrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/postgres_repository"
	"gorm.io/gorm"
)

type PReqService struct {
	PRRepo   *postgresrepository.PReqRepository
	TeamRepo *postgresrepository.TeamRepository
	UserRepo *postgresrepository.UserRepository
}

func NewPReqService(prRepo *postgresrepository.PReqRepository, teamRepo *postgresrepository.TeamRepository, userRepo *postgresrepository.UserRepository) *PReqService {
	return &PReqService{
		PRRepo:   prRepo,
		TeamRepo: teamRepo,
		UserRepo: userRepo,
	}
}

func (prserv *PReqService) CreatePullRequest(ctx context.Context, prReqBody openapi.PostPullRequestCreateJSONBody) (*openapi.PullRequest, *serviceerrors.ServiceError) {
	author, err := prserv.UserRepo.GetUserByCustomId(ctx, prReqBody.AuthorId)
	if err != nil {
		return nil, serviceerrors.ErrUnknown
	}
	if author == nil {
		return nil, serviceerrors.ErrUserNotFound
	}

	var team *models.Team
	if author.TeamID != nil {
		team, err = prserv.TeamRepo.GetTeamByID(ctx, *author.TeamID)
	}
	if err != nil {
		return nil, serviceerrors.ErrUnknown
	}
	if team == nil {
		return nil, serviceerrors.ErrTeamNotFound
	}

	candidates, err := prserv.TeamRepo.GetAllParticipantsButNotSpecial(ctx, team.ID.String(), author.ID.String())
	if err != nil {
		return nil, serviceerrors.ErrUnknown
	}

	reviewers := make([]*models.User, 0, 2)

	if len(candidates) >= 3 {
		ind_1 := rand.Intn(len(candidates))
		ind_2 := rand.Intn(len(candidates) - 1)
		if ind_1 == ind_2 {
			ind_2++
		}
		reviewers = append(reviewers, candidates[ind_1], candidates[ind_2])
	} else {
		reviewers = candidates
	}

	pr := &models.PullRequest{
		PullRequestCustomID: prReqBody.PullRequestId,
		PullRequestName:     prReqBody.PullRequestName,
		AuthorID:            author.ID,
		Status:              "OPEN",
		AssignedReviewers:   reviewers,
	}
	if err := prserv.PRRepo.CreatePullRequest(ctx, pr); err != nil {
		if err == postgresrepository.ErrPRExists {
			return nil, serviceerrors.ErrPRExists
		}
		return nil, serviceerrors.ErrUnknown
	}

	resp := &openapi.PullRequest{
		PullRequestId:     pr.PullRequestCustomID,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          author.UserCustomID,
		Status:            openapi.PullRequestStatus(pr.Status),
		AssignedReviewers: make([]string, 0, len(pr.AssignedReviewers)),
	}
	for _, r := range pr.AssignedReviewers {
		resp.AssignedReviewers = append(resp.AssignedReviewers, r.UserCustomID)
	}
	return resp, nil
}

func (prserv *PReqService) MarkPullReqAsMerged(ctx context.Context, prId string) (*openapi.PullRequest, *serviceerrors.ServiceError) {
	pullRequest, err := prserv.PRRepo.GetPullRequestByID(ctx, prId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.ErrPRNotFound
		}
		return nil, serviceerrors.ErrUnknown
	}

	if pullRequest == nil {
		return nil, serviceerrors.ErrPRNotFound
	}

	if pullRequest.Status != "MERGED" {
		pullRequest.Status = "MERGED"
		now := time.Now().Unix()
		pullRequest.MergedAt = &now
		if err := prserv.PRRepo.UpdatePullRequest(ctx, pullRequest); err != nil {
			return nil, serviceerrors.ErrUnknown
		}
	}

	crAt := time.Unix(pullRequest.CreatedAt, 0)
	var mrAt *time.Time
	if pullRequest.MergedAt != nil {
		t := time.Unix(*pullRequest.MergedAt, 0)
		mrAt = &t
	}

	resp := &openapi.PullRequest{
		AuthorId:          pullRequest.Author.UserCustomID,
		CreatedAt:         &crAt,
		MergedAt:          mrAt,
		PullRequestId:     pullRequest.PullRequestCustomID,
		PullRequestName:   pullRequest.PullRequestName,
		Status:            openapi.PullRequestStatus(pullRequest.Status),
		AssignedReviewers: make([]string, 0, len(pullRequest.AssignedReviewers)),
	}
	for _, r := range pullRequest.AssignedReviewers {
		resp.AssignedReviewers = append(resp.AssignedReviewers, r.UserCustomID)
	}
	return resp, nil
}

func (prserv *PReqService) ReassignReviewer(ctx context.Context, prId, old_reviewer_id string) (*models.PullRequestReassign, *serviceerrors.ServiceError) {
	pullRequest, err := prserv.PRRepo.GetPullRequestByID(ctx, prId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.ErrPRNotFound
		}
		return nil, serviceerrors.ErrUnknown
	}

	if pullRequest == nil {
		return nil, serviceerrors.ErrPRNotFound
	}

	if pullRequest.Status != "OPEN" {
		return nil, serviceerrors.ErrPRMerged
	}

	var oldReviewer *models.User
	var oldIndex int

	for i, reviewer := range pullRequest.AssignedReviewers {
		if reviewer.UserCustomID == old_reviewer_id {
			oldReviewer = reviewer
			oldIndex = i
			break
		}
	}
	if oldReviewer == nil {
		return nil, serviceerrors.ErrNotAssigned
	}
	var team *models.Team
	if oldReviewer.TeamID != nil {
		team, err = prserv.TeamRepo.GetTeamByID(ctx, *oldReviewer.TeamID)
	}
	if err != nil {
		return nil, serviceerrors.ErrUnknown
	}
	if team == nil {
		return nil, serviceerrors.ErrUnknown
	}

	excluded := make([]*models.User, 0, len(pullRequest.AssignedReviewers)+1)
	excluded = append(excluded, pullRequest.AssignedReviewers...)
	excluded = append(excluded, &pullRequest.Author)

	newReviewer := prserv.TeamRepo.PickMemberNotInList(team.Members, excluded)
	if newReviewer == nil {
		return nil, serviceerrors.ErrNoCandidate
	}

	pullRequest.AssignedReviewers[oldIndex] = newReviewer
	if err := prserv.PRRepo.UpdatePullRequest(ctx, pullRequest); err != nil {
		return nil, serviceerrors.ErrUnknown
	}

	crAt := time.Unix(pullRequest.CreatedAt, 0)
	var mrAt *time.Time
	if pullRequest.MergedAt != nil {
		t := time.Unix(*pullRequest.MergedAt, 0)
		mrAt = &t
	}

	resp := models.PullRequestReassign{
		PullRequest: openapi.PullRequest{
			AssignedReviewers: make([]string, 0, len(pullRequest.AssignedReviewers)),
			AuthorId:          pullRequest.Author.UserCustomID,
			CreatedAt:         &crAt,
			MergedAt:          mrAt,
			PullRequestId:     pullRequest.PullRequestCustomID,
			PullRequestName:   pullRequest.PullRequestName,
			Status:            openapi.PullRequestStatus(pullRequest.Status),
		},
		NewReviewerID: newReviewer.UserCustomID,
	}
	for _, r := range pullRequest.AssignedReviewers {
		resp.PullRequest.AssignedReviewers = append(resp.PullRequest.AssignedReviewers, r.UserCustomID)
	}
	return &resp, nil
}

func (prserv *PReqService) GetPullReqsByReviever(ctx context.Context, reviewer_id string) (*models.PullRequestSearch, *serviceerrors.ServiceError) {
	prList, err := prserv.PRRepo.ListPullRequestsByReviewerCustomID(ctx, reviewer_id)
	if err != nil {
		return nil, serviceerrors.ErrUnknown
	}

	resp := &models.PullRequestSearch{
		PullRequest: make([]*openapi.PullRequestShort, 0, len(prList)),
		Author:      reviewer_id,
	}
	for _, pr := range prList {
		resp.PullRequest = append(resp.PullRequest, &openapi.PullRequestShort{
			AuthorId:        pr.Author.UserCustomID,
			PullRequestId:   pr.PullRequestCustomID,
			PullRequestName: pr.PullRequestName,
			Status:          openapi.PullRequestShortStatus(pr.Status),
		})
	}

	return resp, nil

}
