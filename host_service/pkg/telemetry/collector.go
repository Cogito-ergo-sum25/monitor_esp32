package telemetry

import (
	"monitor-esp32-host/pkg/models"
)

// Collector orquesta la lectura de CPU, GPU y RAM en cualquier sistema operativo.
type Collector struct {
	gpuMonitor GPUMonitor
}

func NewCollector() *Collector {
	// Llamada inicial para calibrar el cálculo de porcentaje de CPU
	GetCPUMetrics()

	return &Collector{
		gpuMonitor: NewPlatformGPUMonitor(),
	}
}

// GetPayload devuelve el payload completo de telemetría listo para ser serializado a JSON.
func (c *Collector) GetPayload() models.TelemetryPayload {
	gpuMetrics, _ := c.gpuMonitor.GetMetrics()

	return models.TelemetryPayload{
		CPU: GetCPUMetrics(),
		GPU: gpuMetrics,
		RAM: GetRAMMetrics(),
	}
}
