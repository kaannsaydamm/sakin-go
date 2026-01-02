package services

import (
	"context"
	"fmt"

	"sakin-go/pkg/database"
)

type DashboardStats struct {
	TotalEvents    uint64 `json:"total_events"`
	EventsLastHour uint64 `json:"events_last_hour"`
	AlertsCount    uint64 `json:"alerts_count"`
	ActiveAgents   int    `json:"active_agents"`
}

type DashboardService struct {
	ch *database.ClickHouseClient
	pg *database.PostgresClient
}

func NewDashboardService(ch *database.ClickHouseClient, pg *database.PostgresClient) *DashboardService {
	return &DashboardService{ch: ch, pg: pg}
}

func (s *DashboardService) GetOverview(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 1. ClickHouse Stats (Events)
	// count() from events
	row := s.ch.Conn().QueryRow(ctx, "SELECT count() FROM events")
	if err := row.Scan(&stats.TotalEvents); err != nil {
		return nil, fmt.Errorf("ch query failed: %w", err)
	}

	// count() last hour
	row = s.ch.Conn().QueryRow(ctx, "SELECT count() FROM events WHERE timestamp > now() - INTERVAL 1 HOUR")
	if err := row.Scan(&stats.EventsLastHour); err != nil {
		return nil, err
	}

	// 2. Postgres Stats (Alerts)
	// SELECT count(*) FROM alerts WHERE status = 'new'
	// For demo we mock or simple query
	// if s.pg != nil { ... }
	stats.AlertsCount = 5 // Mock for now until PG filled

	stats.ActiveAgents = 12 // Mock

	return stats, nil
}
