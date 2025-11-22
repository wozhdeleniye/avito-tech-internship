package services

import (
	"context"
	"errors"
	"time"

	"math/rand"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	postgresrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/postgres_repository"
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

func (prserv *PReqService) CreatePullRequest(ctx context.Context, prReqBody openapi.PostPullRequestCreateJSONBody) (*openapi.PullRequest, error) {
	author, err := prserv.UserRepo.GetUserByCustomID(ctx, prReqBody.AuthorId)
	if err != nil || author == nil {
		return nil, err
	}

	team, err := prserv.TeamRepo.GetTeamByID(ctx, author.TeamID)
	if err != nil || team == nil {
		return nil, err
	}

	candidates, err := prserv.TeamRepo.GetAllParticipantsButNotSpecial(ctx, team.ID.String(), author.ID.String())
	if err != nil {
		return nil, err
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
		return nil, err
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

func (prserv *PReqService) MarkPullReqAsMerged(ctx context.Context, prId string) (*openapi.PullRequest, error) {
	pullRequest, err := prserv.PRRepo.GetPullRequestByID(ctx, prId)
	if err != nil {
		return nil, err
	}

	if pullRequest.Status != "MERGED" {
		pullRequest.Status = "MERGED"
		now := time.Now().Unix()
		pullRequest.MergedAt = &now
		err = prserv.PRRepo.UpdatePullRequest(ctx, pullRequest)
		if err != nil {
			return nil, err
		}
	}
	crAt := time.Unix(pullRequest.CreatedAt, 0)
	mrAt := time.Unix(*pullRequest.MergedAt, 0)
	resp := &openapi.PullRequest{
		AuthorId:          pullRequest.Author.UserCustomID,
		CreatedAt:         &crAt,
		MergedAt:          &mrAt,
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

func (prserv *PReqService) ReassignReviewer(ctx context.Context, prId, old_reviewer_id string) (*models.PullRequestReassign, error) {
	pullRequest, err := prserv.PRRepo.GetPullRequestByID(ctx, prId)
	if err != nil {
		return nil, err
	}
	if pullRequest.Status != "OPEN" {
		return nil, errors.New("cannot reassign reviewer for closed pull request")
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
		return nil, errors.New("old reviewer not found")
	}

	team, err := prserv.TeamRepo.GetTeamByID(ctx, oldReviewer.ID)
	if err != nil || team == nil {
		return nil, err
	}
	excluded := append(team.Members, &pullRequest.Author)
	newReviewer := prserv.TeamRepo.PickMemberNotInList(excluded, pullRequest.AssignedReviewers)
	if newReviewer == nil {
		return nil, errors.New("no available new reviewer found")
	}
	pullRequest.AssignedReviewers[oldIndex] = newReviewer
	if err := prserv.PRRepo.UpdatePullRequest(ctx, pullRequest); err != nil {
		return nil, err
	}
	crAt := time.Unix(pullRequest.CreatedAt, 0)
	mrAt := time.Unix(*pullRequest.MergedAt, 0)
	resp := models.PullRequestReassign{
		PullRequest: openapi.PullRequest{
			AssignedReviewers: make([]string, 0, len(pullRequest.AssignedReviewers)),
			AuthorId:          pullRequest.Author.UserCustomID,
			CreatedAt:         &crAt,
			MergedAt:          &mrAt,
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

func (prserv *PReqService) GetPullReqsByReviever(ctx context.Context, reviewer_id string) (*models.PullRequestSearch, error) {
	prList, err := prserv.PRRepo.ListPullRequestsByReviewerCustomID(ctx, reviewer_id)
	if err != nil {
		return nil, err
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
