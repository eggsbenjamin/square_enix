//go:generate mockgen -package repository -source=process.go -destination ./mocks/process.go

package repository

import (
	"errors"
	"log"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
)

var (
	ErrNoProcessExists      = errors.New("no process exists")
	ErrRunningProcessExists = errors.New("running process exists")
)

type ProcessRepository interface {
	CreateNewProcess() (models.Process, error)
	UpdateProcess(models.Process) error
	GetByStatus(status string) ([]models.Process, error)
	GetLatestProcess() (models.Process, error)
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
	defer func() {
		if _, err := p.db.Exec("UNLOCK TABLES"); err != nil {
			log.Fatalf("error unlocking Process table: %q", err)
		}
	}()

	var process models.Process
	if _, err := p.db.Exec("LOCK TABLES Process WRITE"); err != nil {
		return process, err
	}

	runningProcesses, err := p.GetByStatus(models.PROCESS_STATUS_RUNNING)
	if err != nil {
		return process, err
	}

	if len(runningProcesses) > 0 {
		return process, ErrRunningProcessExists
	}

	if _, err := p.db.Exec("INSERT INTO Process (status) VALUES (?)", models.PROCESS_STATUS_RUNNING); err != nil {
		return process, err
	}

	return process, p.db.Get(&process, "SELECT * FROM Process WHERE status = ?", models.PROCESS_STATUS_RUNNING)
}

func (p *processRepo) UpdateProcess(process models.Process) error {
	defer func() {
		if _, err := p.db.Exec("UNLOCK TABLES"); err != nil {
			log.Fatalf("error unlocking Process table: %q", err)
		}
	}()

	if _, err := p.db.Exec("LOCK TABLES Process WRITE"); err != nil {
		return err
	}

	_, err := p.db.Exec(
		"UPDATE Process SET status = ? WHERE id = ?",
		process.Status,
		process.ID,
	)
	return err
}

func (p *processRepo) GetByStatus(status string) ([]models.Process, error) {
	processes := []models.Process{}
	return processes, p.db.Select(&processes, `SELECT * FROM Process WHERE status = ?`, status)
}

func (p *processRepo) GetLatestProcess() (models.Process, error) {
	process := models.Process{}
	if err := p.db.Get(&process, `SELECT * FROM Process ORDER BY created_at DESC LIMIT 1`); err != nil {
		return process, err
	}

	if process.ID == 0 {
		return process, ErrNoProcessExists
	}

	return process, nil
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
