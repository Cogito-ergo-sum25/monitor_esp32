package telemetry

import "monitor-esp32-host/pkg/models"

// MediaMonitor define la interfaz para recolectar información de reproductores multimedia locales.
type MediaMonitor interface {
	GetMediaMetrics() models.MediaMetrics
}
