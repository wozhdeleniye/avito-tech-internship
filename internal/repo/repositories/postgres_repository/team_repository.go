package postgresrepository

import (
	"context"
	"math/rand"
	"strings"

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

// цельная транзакция для создания команды и созд участников
func (r *TeamRepository) CreateTeamWithMembers(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Members").Create(team).Error; err != nil {
			le := strings.ToLower(err.Error())
			if strings.Contains(le, "duplicate") || strings.Contains(le, "unique") || strings.Contains(le, "violates unique") {
				return ErrTeamExists
			}
			return err
		}

		for _, member := range team.Members {
			if member == nil {
				continue
			}

			member.TeamID = &team.ID
			if err := tx.Create(member).Error; err != nil {
				le := strings.ToLower(err.Error())
				if strings.Contains(le, "duplicate") || strings.Contains(le, "unique") || strings.Contains(le, "violates unique") {
					return ErrUserExists
				}
				return err
			}
		}

		if len(team.Members) > 0 {
			toAppend := make([]*models.User, 0, len(team.Members))
			for _, m := range team.Members {
				if m == nil {
					continue
				}
				toAppend = append(toAppend, &models.User{ID: m.ID})
			}

			teamModel := &models.Team{ID: team.ID}
			if err := tx.Model(teamModel).Association("Members").Append(toAppend); err != nil {
				return err
			}
		}

		return nil
	})
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
	var members []*models.User

	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Joins("JOIN user_teams ut ON ut.user_id = users.id").
		Where("ut.team_id = ? AND users.id != ? AND users.is_active = ?", teamID, userID, true).
		Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
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

func (r *TeamRepository) FindTeamByName(ctx context.Context, name string) (*models.Team, error) {
	var team models.Team
	result := r.db.WithContext(ctx).Preload("Members").Where("team_name = ?", name).First(&team)
	if result.Error != nil {
		return nil, result.Error
	}
	return &team, nil
}
