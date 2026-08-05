package memory

import (
	"context"
	"sync"
	"time"
)

// RenderMetric collects render metrics in memory.
type RenderMetric struct {
	mu sync.RWMutex

	totalRenders  int
	totalDuration time.Duration
	totalObjects  int
}

// Observe records a render operation's metrics.
func (m *RenderMetric) Observe(_ context.Context, duration time.Duration, objectCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRenders++
	m.totalDuration += duration
	m.totalObjects += objectCount
}

// Summary returns a snapshot of current render metrics.
func (m *RenderMetric) Summary() RenderSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := time.Duration(0)
	if m.totalRenders > 0 {
		avgDuration = m.totalDuration / time.Duration(m.totalRenders)
	}

	return RenderSummary{
		TotalRenders:    m.totalRenders,
		AverageDuration: avgDuration,
		TotalObjects:    m.totalObjects,
	}
}

// RenderSummary provides a snapshot of render metrics.
type RenderSummary struct {
	TotalRenders    int
	AverageDuration time.Duration
	TotalObjects    int
}

// RendererMetric collects renderer-specific metrics in memory.
type RendererMetric struct {
	mu        sync.RWMutex
	renderers map[string]*rendererStats
}

type rendererStats struct {
	executions int
	duration   time.Duration
	objects    int
	errors     int
}

// NewRendererMetric creates a new renderer metrics collector.
func NewRendererMetric() *RendererMetric {
	return &RendererMetric{
		renderers: make(map[string]*rendererStats),
	}
}

// Observe records a renderer execution's metrics.
func (m *RendererMetric) Observe(
	_ context.Context,
	rendererType string,
	duration time.Duration,
	objectCount int,
	err error,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.renderers[rendererType]; !exists {
		m.renderers[rendererType] = &rendererStats{}
	}

	stats := m.renderers[rendererType]
	stats.executions++
	stats.duration += duration
	stats.objects += objectCount
	if err != nil {
		stats.errors++
	}
}

// Summary returns a snapshot of current renderer metrics.
func (m *RendererMetric) Summary() map[string]RendererSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]RendererSummary)
	for name, stats := range m.renderers {
		avgDuration := time.Duration(0)
		if stats.executions > 0 {
			avgDuration = stats.duration / time.Duration(stats.executions)
		}

		result[name] = RendererSummary{
			Executions:      stats.executions,
			AverageDuration: avgDuration,
			TotalObjects:    stats.objects,
			Errors:          stats.errors,
		}
	}

	return result
}

// RendererSummary provides a snapshot of metrics for a specific renderer.
type RendererSummary struct {
	Executions      int
	AverageDuration time.Duration
	TotalObjects    int
	Errors          int
}
