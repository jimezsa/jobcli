package scraper

import (
	"context"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
)

type GoogleJobs struct {
	client *network.Client
}

func NewGoogleJobs(client *network.Client) *GoogleJobs {
	return &GoogleJobs{client: client}
}

func (g *GoogleJobs) Name() string {
	return SiteGoogleJobs
}

func (g *GoogleJobs) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	return nil, ErrNotImplemented
}
