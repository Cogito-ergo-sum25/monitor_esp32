package telemetry

import (
	"monitor-esp32-host/pkg/models"
)

// Collector orquesta la lectura de CPU, GPU, RAM y Media en cualquier sistema operativo.
type Collector struct {
	gpuMonitor   GPUMonitor
	mediaMonitor MediaMonitor
}

func NewCollector() *Collector {
	// Llamada inicial para calibrar el cálculo de porcentaje de CPU
	GetCPUMetrics()

	return &Collector{
		gpuMonitor:   NewPlatformGPUMonitor(),
		mediaMonitor: NewPlatformMediaMonitor(),
	}
}

// GetPayload devuelve el payload completo de telemetría listo para ser serializado a JSON.
func (c *Collector) GetPayload() models.TelemetryPayload {
	gpuMetrics, _ := c.gpuMonitor.GetMetrics()

	return models.TelemetryPayload{
		CPU:   GetCPUMetrics(),
		GPU:   gpuMetrics,
		RAM:   GetRAMMetrics(),
		Media: c.mediaMonitor.GetMediaMetrics(),
	}
}
