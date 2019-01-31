package repository

import (
	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
)

type ProcessRepository interface {
	CreateNewProcess() (models.Process, error)
	UpdateProcess(models.Process) error
	GetByStatus(status string) ([]models.Process, error)
}

type processRepo struct {
	db db.Querier
}

func NewProcessRepository(db db.Querier) ProcessRepository {
	return &processRepo{
		db: db,
	}
}

func (p *processRepo) CreateNewProcess() (models.Process, error) {
	var process models.Process
	if _, err := p.db.Exec("LOCK TABLES Process WRITE"); err != nil {
		return process, err
	}

	if _, err := p.db.Exec("INSERT INTO Process (status) VALUES (?)", models.PROCESS_STATUS_RUNNING); err != nil {
		return process, err
	}

	if err := p.db.Get(&process, "SELECT * FROM Process WHERE status = ?", models.PROCESS_STATUS_RUNNING); err != nil {
		return process, err
	}

	if _, err := p.db.Exec("UNLOCK TABLES"); err != nil {
		return process, err
	}

	return process, nil
}

func (p *processRepo) UpdateProcess(process models.Process) error {
	if _, err := p.db.Exec("LOCK TABLES Process WRITE"); err != nil {
		return err
	}

	if _, err := p.db.Exec(
		"UPDATE Process SET status = ? WHERE id = ?",
		process.Status,
		process.ID,
	); err != nil {
		return err
	}

	_, err := p.db.Exec("UNLOCK TABLES")
	return err
}

func (p *processRepo) GetByStatus(status string) ([]models.Process, error) {
	processes := []models.Process{}
	return processes, p.db.Select(&processes, `SELECT * FROM Process WHERE status = ?`, status)
}

type ProcessRepositoryFactory interface {
	CreateProcessRepository(db db.Querier) ProcessRepository
}

type processRepoFactory struct{}

func NewProcessRepositoryFactory() ProcessRepositoryFactory {
	return &processRepoFactory{}
}

func (p *processRepoFactory) CreateProcessRepository(db db.Querier) ProcessRepository {
	return NewProcessRepository(db)
}
