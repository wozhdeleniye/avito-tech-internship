package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	TeamName string    `gorm:"unique;not null" json:"team_name"`
	Members  []*User   `gorm:"many2many:user_teams;" json:"members"`
}

func (p *Team) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
