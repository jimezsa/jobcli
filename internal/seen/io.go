package seen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jimezsa/jobcli/internal/models"
)

// ReadJobs reads a JSON array of jobs from path.
func ReadJobs(path string) ([]models.Job, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []models.Job{}, nil
	}

	var jobs []models.Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, err
	}
	if jobs == nil {
		return []models.Job{}, nil
	}
	return jobs, nil
}

// ReadJobsAllowMissing reads jobs and treats missing files as empty history.
func ReadJobsAllowMissing(path string) ([]models.Job, error) {
	jobs, err := ReadJobs(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []models.Job{}, nil
		}
		return nil, err
	}
	return jobs, nil
}

// WriteJobs writes jobs as pretty JSON.
func WriteJobs(path string, jobs []models.Job) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path is required")
	}
	if jobs == nil {
		jobs = []models.Job{}
	}
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
