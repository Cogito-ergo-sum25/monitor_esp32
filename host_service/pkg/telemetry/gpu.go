package telemetry

import "monitor-esp32-host/pkg/models"

// GPUMonitor define la interfaz que cualquier proveedor de métricas de GPU debe implementar.
type GPUMonitor interface {
	GetMetrics() (models.GPUMetrics, error)
}
