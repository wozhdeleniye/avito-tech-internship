package services

import (
	"context"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
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

func (ts *TeamService) CreateTeam(ctx context.Context, req openapi.Team) (*openapi.Team, error) {
	newTeam := models.Team{
		TeamName: req.TeamName,
		Members:  make([]*models.User, 0, len(req.Members)),
	}
	for _, member := range req.Members {
		user, err := ts.UserRepo.GetUserByCustomID(ctx, member.UserId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, ErrUserNotFound //поправить ошибки
		}
		newTeam.Members = append(newTeam.Members, user)
	}
	ts.TeamRepo.CreateTeam(ctx, &newTeam)
	return &req, nil
}

func (ts *TeamService) GetTeamQuery(ctx context.Context, req openapi.TeamNameQuery) (*openapi.Team, error) {
	team, err := ts.TeamRepo.FindTeamsByName(ctx, req)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrUserNotFound //поправить ошибки
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
