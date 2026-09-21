package serialutil

import (
	"strings"

	"go.bug.st/serial"
)

// FindESP32Port escanea los puertos serie del sistema y selecciona el candidato más probable.
func FindESP32Port() string {
	ports, err := serial.GetPortsList()
	if err != nil || len(ports) == 0 {
		return ""
	}

	// 1. Priorizar puertos comunes en Linux y macOS
	for _, port := range ports {
		p := strings.ToLower(port)
		if strings.Contains(p, "ttyusb") || strings.Contains(p, "ttyacm") ||
			strings.Contains(p, "usbserial") || strings.Contains(p, "usbmodem") {
			return port
		}
	}

	// 2. En Windows, los puertos son COM1, COM3, COM4...
	// COM1 suele ser el puerto heredado de la placa base; priorizamos COM3 en adelante.
	for _, port := range ports {
		p := strings.ToUpper(port)
		if strings.HasPrefix(p, "COM") && p != "COM1" {
			return port
		}
	}

	// 3. Fallback al primer puerto si solo hay uno disponible
	return ports[0]
}
