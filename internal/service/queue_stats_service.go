package service

import (
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
)

// QueueStatsItemPayload is the API-facing shape for one queue stats row.
type QueueStatsItemPayload struct {
	QueueKey   string `json:"queue_key"`
	Success    int64  `json:"success"`
	Failed     int64  `json:"failed"`
	Processing int64  `json:"processing"`
	Pending    int64  `json:"pending"`
	Submitted  int64  `json:"submitted"`
}

// QueueStatsService provides queue dashboard stats.
type QueueStatsService interface {
	GetAll() ([]QueueStatsItemPayload, error)
}

type queueStatsService struct {
	jobs repository.QueueJobRepository
}

// NewQueueStatsService creates a queue stats service.
func NewQueueStatsService(jobs repository.QueueJobRepository) QueueStatsService {
	return &queueStatsService{jobs: jobs}
}

// GetAll returns queue stats from the unified queue job table.
func (s *queueStatsService) GetAll() ([]QueueStatsItemPayload, error) {
	rows, err := s.jobs.ListStatusCounts()
	if err != nil {
		return nil, err
	}

	statsByKey := make(map[string]QueueStatsItemPayload, len(queue.Definitions()))
	for _, definition := range queue.Definitions() {
		statsByKey[string(definition.Key)] = QueueStatsItemPayload{
			QueueKey: string(definition.Key),
		}
	}

	for _, row := range rows {
		item := statsByKey[row.QueueKey]
		item.Submitted += row.Count
		switch row.Status {
		case model.QueueJobStatusCompleted:
			item.Success += row.Count
		case model.QueueJobStatusFailed:
			item.Failed += row.Count
		case model.QueueJobStatusProcessing:
			item.Processing += row.Count
		case model.QueueJobStatusPending:
			item.Pending += row.Count
		}
		statsByKey[row.QueueKey] = item
	}

	items := make([]QueueStatsItemPayload, 0, len(queue.Definitions()))
	for _, definition := range queue.Definitions() {
		items = append(items, statsByKey[string(definition.Key)])
	}

	return items, nil
}
