//go:build linux

package telemetry

import (
	"strings"

	"github.com/godbus/dbus/v5"

	"monitor-esp32-host/pkg/models"
)

type LinuxMediaMonitor struct {
	conn *dbus.Conn
}

func NewPlatformMediaMonitor() MediaMonitor {
	conn, _ := dbus.SessionBus()
	return &LinuxMediaMonitor{conn: conn}
}

func (m *LinuxMediaMonitor) GetMediaMetrics() models.MediaMetrics {
	var metrics models.MediaMetrics

	if m.conn == nil {
		conn, err := dbus.SessionBus()
		if err != nil {
			return metrics
		}
		m.conn = conn
	}

	// 1. Listar nombres en el Session Bus
	var names []string
	err := m.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return metrics
	}

	// 2. Buscar reproductores MPRIS disponibles (Priorizando Spotify)
	var chosenPlayer string
	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			if strings.Contains(name, "spotify") {
				chosenPlayer = name
				break
			}
			if chosenPlayer == "" {
				chosenPlayer = name
			}
		}
	}

	if chosenPlayer == "" {
		return metrics
	}

	obj := m.conn.Object(chosenPlayer, "/org/mpris/MediaPlayer2")

	// 3. PlaybackStatus (Playing / Paused / Stopped)
	statusProp, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err == nil {
		if s, ok := statusProp.Value().(string); ok {
			metrics.IsPlaying = (s == "Playing")
		}
	}

	// 4. Position (en microsegundos en MPRIS)
	posProp, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Position")
	if err == nil {
		if p, ok := posProp.Value().(int64); ok {
			metrics.ProgressMs = int(p / 1000)
		}
	}

	// 5. Metadata
	metaProp, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err == nil {
		if metaMap, ok := metaProp.Value().(map[string]dbus.Variant); ok {
			// Título
			if titleVar, exists := metaMap["xesam:title"]; exists {
				if t, ok := titleVar.Value().(string); ok {
					metrics.Title = t
				}
			}

			// Artistas (puede ser []string o string)
			if artistVar, exists := metaMap["xesam:artist"]; exists {
				if artists, ok := artistVar.Value().([]string); ok && len(artists) > 0 {
					metrics.Artist = strings.Join(artists, ", ")
				} else if artistStr, ok := artistVar.Value().(string); ok {
					metrics.Artist = artistStr
				}
			}

			// Duración (mpris:length en microsegundos)
			if lenVar, exists := metaMap["mpris:length"]; exists {
				if l, ok := lenVar.Value().(uint64); ok {
					metrics.DurationMs = int(l / 1000)
				} else if l, ok := lenVar.Value().(int64); ok {
					metrics.DurationMs = int(l / 1000)
				}
			}
		}
	}

	return metrics
}
