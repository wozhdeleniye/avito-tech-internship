package migrations

import (
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	"gorm.io/gorm"
)

type GormMigrator struct {
	db *gorm.DB
}

func NewGormMigrator(db *gorm.DB) *GormMigrator {
	return &GormMigrator{db: db}
}

func (m *GormMigrator) Migrate() error {
	m.db.Exec("SET CONSTRAINTS ALL DEFERRED")

	tables, err := m.db.Migrator().GetTables()
	if err != nil {
		return err
	}

	// закомментировать чтобы не дропать таблицы при миграции(если не забуду добавлю флажок для dev режима в конфиг)
	for _, table := range tables {
		if err := m.db.Migrator().DropTable(table); err != nil {
			return err
		}
	}

	err = m.db.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.PullRequest{},
	)
	return err
}
