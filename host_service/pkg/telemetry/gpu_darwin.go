//go:build darwin || (!linux && !windows)

package telemetry

import (
	"monitor-esp32-host/pkg/models"
)

type FallbackGPUMonitor struct{}

func NewPlatformGPUMonitor() GPUMonitor {
	return &FallbackGPUMonitor{}
}

func (m *FallbackGPUMonitor) GetMetrics() (models.GPUMetrics, error) {
	return models.GPUMetrics{
		Usage:     0.0,
		Temp:      0.0,
		VramUsed:  0.0,
		VramTotal: 0.0,
	}, nil
}
