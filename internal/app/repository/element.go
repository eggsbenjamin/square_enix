//go:generate mockgen -package repository -source=element.go -destination ./mocks/element.go

package repository

import (
	"time"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
)

type ElementRepository interface {
	UpdateElementForProcess(element models.Element, processID int) error
	LockElementsForUpdate(processID int, batchSize int) ([]models.Element, error)
	GetElementsByProcessID(processID int) ([]models.Element, error)
	GetElementsCreatedBefore(date time.Time) ([]models.Element, error)
}

type elementRepo struct {
	db db.Querier
}

func NewElementRepository(db db.Querier) ElementRepository {
	return &elementRepo{
		db: db,
	}
}

func (p *elementRepo) UpdateElementForProcess(element models.Element, processID int) error {
	if _, err := p.db.Exec(
		"UPDATE Element SET data = ? WHERE id = ?",
		element.Data,
		element.ID,
	); err != nil {
		return err
	}

	_, err := p.db.Exec(
		"INSERT INTO ProcessElement (process_id, element_id) VALUES (?, ?)",
		processID,
		element.ID,
	)
	return err
}

func (e *elementRepo) LockElementsForUpdate(processID int, batchSize int) ([]models.Element, error) {
	elements := []models.Element{}
	return elements, e.db.Select(&elements, `
		SELECT e.* FROM Element AS e
		WHERE
		NOT EXISTS (
			SELECT * FROM ProcessElement
			WHERE process_id = ? AND element_id = e.id
		)
		AND
			e.created_at < (SELECT created_at FROM Process WHERE id = ?)
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`,
		processID,
		processID,
		batchSize,
	)
}

func (e *elementRepo) GetElementsByProcessID(processID int) ([]models.Element, error) {
	elements := []models.Element{}
	return elements, e.db.Select(
		&elements,
		`
			SELECT e.id, e.data, e.created_at FROM Element AS e
				INNER JOIN ProcessElement AS pe ON e.id = pe.element_id
			WHERE pe.process_id = ?
		`,
		processID,
	)
}

func (e *elementRepo) GetElementsCreatedBefore(date time.Time) ([]models.Element, error) {
	elements := []models.Element{}
	return elements, e.db.Select(
		&elements,
		`
			SELECT e.* FROM Element AS e
			WHERE e.created_at < ?
		`,
		date,
	)
}

type ElementRepositoryFactory interface {
	CreateElementRepository(db db.Querier) ElementRepository
}

type elementRepoFactory struct{}

func NewElementRepositoryFactory() ElementRepositoryFactory {
	return &elementRepoFactory{}
}

func (p *elementRepoFactory) CreateElementRepository(db db.Querier) ElementRepository {
	return NewElementRepository(db)
}
