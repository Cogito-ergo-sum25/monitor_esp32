package telemetry

import (
	"math"
	"strings"

	"monitor-esp32-host/pkg/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

// GetCPUMetrics recolecta el porcentaje de uso de la CPU y la temperatura del procesador.
func GetCPUMetrics() models.CPUMetrics {
	var metrics models.CPUMetrics

	// 1. Porcentaje de uso total de CPU (no bloqueante)
	percentages, err := cpu.Percent(0, false)
	if err == nil && len(percentages) > 0 {
		metrics.Usage = math.Round(percentages[0]*10) / 10
	}

	// 2. Sensores de temperatura
	temps, err := host.SensorsTemperatures()
	if err == nil {
		// Prioridad 1: k10temp (AMD Ryzen), coretemp (Intel), zenpower
		for _, t := range temps {
			key := strings.ToLower(t.SensorKey)
			if strings.Contains(key, "k10temp") || strings.Contains(key, "coretemp") || strings.Contains(key, "zenpower") {
				if t.Temperature > 0 {
					metrics.Temp = math.Round(t.Temperature*10) / 10
					break
				}
			}
		}

		// Prioridad 2: otros sensores que contengan "cpu" o "acpitz"
		if metrics.Temp == 0 {
			for _, t := range temps {
				key := strings.ToLower(t.SensorKey)
				if strings.Contains(key, "cpu") || strings.Contains(key, "acpitz") {
					if t.Temperature > 0 {
						metrics.Temp = math.Round(t.Temperature*10) / 10
						break
					}
				}
			}
		}
	}

	return metrics
}
