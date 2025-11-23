package services

import (
	"context"
	"errors"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	serviceerrors "github.com/wozhdeleniye/avito-tech-internship/internal/pkg/errors"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	postgresrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/postgres_repository"
)

type TeamService struct {
	PRRepo   *postgresrepository.PReqRepository
	TeamRepo *postgresrepository.TeamRepository
	UserRepo *postgresrepository.UserRepository
}

func NewTeamService(prRepo *postgresrepository.PReqRepository, teamRepo *postgresrepository.TeamRepository, userRepo *postgresrepository.UserRepository) *TeamService {
	return &TeamService{
		PRRepo:   prRepo,
		TeamRepo: teamRepo,
		UserRepo: userRepo,
	}
}

func (ts *TeamService) CreateTeam(ctx context.Context, req openapi.Team) (*openapi.Team, *serviceerrors.ServiceError) {
	newTeam := models.Team{
		TeamName: req.TeamName,
		Members:  make([]*models.User, 0, len(req.Members)),
	}
	for _, member := range req.Members {
		if member.UserId == "" {
			continue
		}

		user := &models.User{
			Nickname:     member.Username,
			UserCustomID: member.UserId,
			IsActive:     member.IsActive,
		}

		newTeam.Members = append(newTeam.Members, user)
	}

	if err := ts.TeamRepo.CreateTeamWithMembers(ctx, &newTeam); err != nil {
		if errors.Is(err, postgresrepository.ErrUserExists) {
			return nil, serviceerrors.ErrUserExists
		}
		if errors.Is(err, postgresrepository.ErrTeamExists) {
			return nil, serviceerrors.ErrTeamExists
		}
		return nil, serviceerrors.ErrUnknown
	}

	teamResp := openapi.Team{
		TeamName: newTeam.TeamName,
		Members:  make([]openapi.TeamMember, 0, len(newTeam.Members)),
	}
	for _, member := range newTeam.Members {
		teamResp.Members = append(teamResp.Members, openapi.TeamMember{
			IsActive: member.IsActive,
			UserId:   member.UserCustomID,
			Username: member.Nickname,
		})
	}
	return &teamResp, nil
}

func (ts *TeamService) GetTeamQuery(ctx context.Context, req openapi.TeamNameQuery) (*openapi.Team, *serviceerrors.ServiceError) {
	team, err := ts.TeamRepo.FindTeamByName(ctx, req)
	if err != nil {
		return nil, serviceerrors.ErrTeamNotFound
	}
	if team == nil {
		return nil, serviceerrors.ErrTeamNotFound
	}

	teamResp := openapi.Team{
		TeamName: team.TeamName,
		Members:  make([]openapi.TeamMember, 0, len(team.Members)),
	}
	for _, member := range team.Members {
		teamResp.Members = append(teamResp.Members, openapi.TeamMember{
			IsActive: member.IsActive,
			UserId:   member.UserCustomID,
			Username: member.Nickname,
		})
	}
	return &teamResp, nil
}

func (s *TeamService) SetUserActive(ctx context.Context, userId string, isActive bool) (*models.User, *serviceerrors.ServiceError) {
	user, err := s.UserRepo.GetUserByCustomId(ctx, userId)
	if err != nil {
		return nil, serviceerrors.ErrUnknown
	}
	if user == nil {
		return nil, serviceerrors.ErrUserNotFound
	}

	user.IsActive = isActive

	if err := s.UserRepo.UpdateUser(ctx, user); err != nil {
		return nil, serviceerrors.ErrUnknown
	}

	return user, nil
}
