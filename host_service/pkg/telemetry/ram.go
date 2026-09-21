package telemetry

import (
	"math"

	"monitor-esp32-host/pkg/models"

	"github.com/shirou/gopsutil/v3/mem"
)

// GetRAMMetrics recolecta el porcentaje de memoria RAM en uso, así como los GB usados y totales.
func GetRAMMetrics() models.RAMMetrics {
	var metrics models.RAMMetrics

	vm, err := mem.VirtualMemory()
	if err == nil {
		metrics.Usage = math.Round(vm.UsedPercent*10) / 10
		metrics.Used = math.Round((float64(vm.Used)/(1024*1024*1024))*10) / 10
		metrics.Total = math.Round((float64(vm.Total)/(1024*1024*1024))*10) / 10
	}

	return metrics
}
