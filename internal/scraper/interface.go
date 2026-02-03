package scraper

import (
	"context"
	"errors"

	"github.com/MrJJimenez/jobcli/internal/models"
)

var ErrNotImplemented = errors.New("scraper not implemented")

type Scraper interface {
	Name() string
	Search(ctx context.Context, params models.SearchParams) ([]models.Job, error)
}
