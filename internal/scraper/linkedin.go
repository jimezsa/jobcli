package scraper

import (
	"context"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
)

type LinkedIn struct {
	client *network.Client
}

func NewLinkedIn(client *network.Client) *LinkedIn {
	return &LinkedIn{client: client}
}

func (l *LinkedIn) Name() string {
	return SiteLinkedIn
}

func (l *LinkedIn) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	return nil, ErrNotImplemented
}
