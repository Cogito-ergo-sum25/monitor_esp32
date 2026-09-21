//go:build !linux

package telemetry

import "monitor-esp32-host/pkg/models"

type FallbackMediaMonitor struct{}

func NewPlatformMediaMonitor() MediaMonitor {
	return &FallbackMediaMonitor{}
}

func (m *FallbackMediaMonitor) GetMediaMetrics() models.MediaMetrics {
	return models.MediaMetrics{
		Title:      "Desk Dashboard",
		Artist:     "Media no disponible",
		IsPlaying:  false,
		ProgressMs: 0,
		DurationMs: 1000,
	}
}
