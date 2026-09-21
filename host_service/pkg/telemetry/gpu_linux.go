//go:build linux

package telemetry

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"monitor-esp32-host/pkg/models"

	"github.com/shirou/gopsutil/v3/host"
)

type LinuxGPUMonitor struct {
	cardPath string
}

func NewPlatformGPUMonitor() GPUMonitor {
	// Buscar ruta de GPU en /sys/class/drm/card*/device
	matches, _ := filepath.Glob("/sys/class/drm/card[0-9]*/device/gpu_busy_percent")
	var foundPath string
	if len(matches) > 0 {
		foundPath = filepath.Dir(matches[0])
	}

	return &LinuxGPUMonitor{
		cardPath: foundPath,
	}
}

func (m *LinuxGPUMonitor) GetMetrics() (models.GPUMetrics, error) {
	var metrics models.GPUMetrics

	// 1. Lectura de uso y VRAM vía sysfs de Linux
	if m.cardPath != "" {
		// Uso de GPU
		busyFile := filepath.Join(m.cardPath, "gpu_busy_percent")
		if data, err := os.ReadFile(busyFile); err == nil {
			if val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
				metrics.Usage = math.Round(val*10) / 10
			}
		}

		// VRAM Usada
		vramUsedFile := filepath.Join(m.cardPath, "mem_info_vram_used")
		if data, err := os.ReadFile(vramUsedFile); err == nil {
			if bytesVal, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
				metrics.VramUsed = math.Round((bytesVal/(1024*1024*1024))*10) / 10
			}
		}

		// VRAM Total
		vramTotalFile := filepath.Join(m.cardPath, "mem_info_vram_total")
		if data, err := os.ReadFile(vramTotalFile); err == nil {
			if bytesVal, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
				metrics.VramTotal = math.Round((bytesVal/(1024*1024*1024))*10) / 10
			}
		}

		// Temperatura vía hwmon
		hwmonMatches, _ := filepath.Glob(filepath.Join(m.cardPath, "hwmon", "hwmon*", "temp1_input"))
		if len(hwmonMatches) > 0 {
			if data, err := os.ReadFile(hwmonMatches[0]); err == nil {
				if milliDeg, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
					metrics.Temp = math.Round((milliDeg/1000.0)*10) / 10
				}
			}
		}
	}

	// 2. Fallback de temperatura si no se leyó de hwmon
	if metrics.Temp == 0 {
		if temps, err := host.SensorsTemperatures(); err == nil {
			for _, t := range temps {
				sensorKey := strings.ToLower(t.SensorKey)
				if strings.Contains(sensorKey, "amdgpu") || strings.Contains(sensorKey, "nvidia") {
					metrics.Temp = math.Round(t.Temperature*10) / 10
					break
				}
			}
		}
	}

	return metrics, nil
}
