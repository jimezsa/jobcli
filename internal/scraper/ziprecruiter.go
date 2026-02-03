package scraper

import (
	"context"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
)

type ZipRecruiter struct {
	client *network.Client
}

func NewZipRecruiter(client *network.Client) *ZipRecruiter {
	return &ZipRecruiter{client: client}
}

func (z *ZipRecruiter) Name() string {
	return SiteZipRecruiter
}

func (z *ZipRecruiter) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	return nil, ErrNotImplemented
}
