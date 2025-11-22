package postgresrepository

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	"gorm.io/gorm"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	result := r.db.WithContext(ctx).Create(team)
	return result.Error
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	var team models.Team
	result := r.db.WithContext(ctx).Preload("Members").Where("id = ?", id).First(&team)
	if result.Error != nil {
		return nil, result.Error
	}
	return &team, nil
}

func (r *TeamRepository) GetAllParticipantsButNotSpecial(ctx context.Context, teamID string, userID string) ([]*models.User, error) {
	var team models.Team
	if err := r.db.WithContext(ctx).Preload("Members", "user_custom_id != ?", userID).Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}
	return team.Members, nil
}

func (r *TeamRepository) PickMemberNotInList(members []*models.User, excluded []*models.User) *models.User {
	if len(members) == 0 {
		return nil
	}

	excludedMap := make(map[string]struct{}, len(excluded))
	for _, e := range excluded {
		if e == nil {
			continue
		}
		excludedMap[e.ID.String()] = struct{}{}
	}

	candidates := make([]*models.User, 0, len(members))
	for _, m := range members {
		if m == nil {
			continue
		}
		if !m.IsActive {
			continue
		}
		if _, ok := excludedMap[m.ID.String()]; ok {
			continue
		}
		candidates = append(candidates, m)
	}

	if len(candidates) == 0 {
		return nil
	}

	return candidates[rand.Intn(len(candidates))]
}

func (r *TeamRepository) FindTeamsByName(ctx context.Context, name string) (*models.Team, error) {
	var team models.Team
	result := r.db.WithContext(ctx).Preload("Members").Where("team_name = ?", name).First(&team)
	if result.Error != nil {
		return nil, result.Error
	}
	return &team, nil
}
