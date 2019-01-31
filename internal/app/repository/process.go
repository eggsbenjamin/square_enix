package repository

import "github.com/eggsbenjamin/square_enix/internal/app/models"

type ProcessRepository interface {
	GetByStatus(status string) ([]models.Process, error)
}
