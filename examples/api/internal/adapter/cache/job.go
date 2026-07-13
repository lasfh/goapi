package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lasfh/goapi/examples/api/internal/core/dto"
	"github.com/redis/go-redis/v9"
)

const (
	jobKeyPrefix = "job:"
	jobTTL       = 1 * time.Hour
)

type jobStatusCache struct {
	rdb *redis.Client
}

func NewJobStatusCache(rdb *redis.Client) *jobStatusCache {
	return &jobStatusCache{rdb: rdb}
}

func (s *jobStatusCache) key(id string) string {
	return jobKeyPrefix + id
}

func (s *jobStatusCache) Create(ctx context.Context) (dto.JobStatus, error) {
	agora := time.Now()
	job := dto.JobStatus{
		ID:           uuid.NewString(),
		Status:       dto.StatusJobPendente,
		CriadoEm:     agora,
		AtualizadoEm: agora,
	}

	if err := s.Save(ctx, job); err != nil {
		return dto.JobStatus{}, err
	}

	return job, nil
}

func (s *jobStatusCache) Find(ctx context.Context, id string) (dto.JobStatus, bool, error) {
	data, err := s.rdb.Get(ctx, s.key(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return dto.JobStatus{}, false, nil
		}

		return dto.JobStatus{}, false, err
	}

	var job dto.JobStatus
	if err := json.Unmarshal(data, &job); err != nil {
		return dto.JobStatus{}, false, err
	}

	return job, true, nil
}

func (s *jobStatusCache) Save(ctx context.Context, job dto.JobStatus) error {
	job.AtualizadoEm = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("serializando job de importação: %w", err)
	}

	return s.rdb.Set(ctx, s.key(job.ID), data, jobTTL).Err()
}
