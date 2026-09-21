//go:build windows

package telemetry

import (
	"monitor-esp32-host/pkg/models"
)

type WindowsGPUMonitor struct{}

func NewPlatformGPUMonitor() GPUMonitor {
	return &WindowsGPUMonitor{}
}

func (m *WindowsGPUMonitor) GetMetrics() (models.GPUMetrics, error) {
	// En Windows, las métricas avanzadas de GPU pueden consultarse mediante
	// WMI, Performance Counters (PDH) o NVML/DirectX.
	// Por defecto inicializamos con valores seguros para compatibilidad total.
	return models.GPUMetrics{
		Usage:     0.0,
		Temp:      0.0,
		VramUsed:  0.0,
		VramTotal: 0.0,
	}, nil
}
