package scraper

import (
	"context"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
)

type Glassdoor struct {
	client *network.Client
}

func NewGlassdoor(client *network.Client) *Glassdoor {
	return &Glassdoor{client: client}
}

func (g *Glassdoor) Name() string {
	return SiteGlassdoor
}

func (g *Glassdoor) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	return nil, ErrNotImplemented
}
