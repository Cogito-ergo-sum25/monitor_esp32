package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.bug.st/serial"

	serialutil "monitor-esp32-host/pkg/serial"
	"monitor-esp32-host/pkg/telemetry"
)

func main() {
	portFlag := flag.String("port", "", "Puerto serie manual (ej. /dev/ttyUSB0 en Linux o COM3 en Windows)")
	baudFlag := flag.Int("baud", 115200, "Velocidad en baudios")
	intervalFlag := flag.Int("interval", 1000, "Intervalo de actualización en milisegundos")
	mockFlag := flag.Bool("mock", false, "Modo de simulación (imprime JSON en consola sin usar puerto serie)")
	flag.Parse()

	collector := telemetry.NewCollector()

	fmt.Println("==================================================")
	fmt.Println("   ESP32 Desk Dashboard - Host Service (Go)       ")
	fmt.Println("==================================================")
	fmt.Printf("Intervalo de actualización: %d ms\n", *intervalFlag)
	fmt.Printf("Baud rate: %d\n", *baudFlag)

	// Manejo de salida limpia con Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n[SALIR] Servicio detenido por el usuario.")
		os.Exit(0)
	}()

	interval := time.Duration(*intervalFlag) * time.Millisecond

	// 1. Modo Simulación
	if *mockFlag {
		fmt.Println("[MODO DE PRUEBA] Simulación activa (--mock). Mostrando JSON en consola.")
		for {
			payload := collector.GetPayload()
			data, err := json.Marshal(payload)
			if err == nil {
				fmt.Printf("[JSON TX] %s\n", string(data))
			}
			time.Sleep(interval)
		}
	}

	// 2. Modo Real con Auto-Reconexión Continua
	mode := &serial.Mode{
		BaudRate: *baudFlag,
	}

	for {
		targetPort := *portFlag
		if targetPort == "" {
			targetPort = serialutil.FindESP32Port()
		}

		if targetPort == "" {
			fmt.Println("[INFO] Esperando conexión de ESP32 (puerto serie no detectado)...")
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("[CONEXIÓN] Intentando abrir puerto: %s a %d baud...\n", targetPort, *baudFlag)
		port, err := serial.Open(targetPort, mode)
		if err != nil {
			fmt.Printf("[ERROR] No se pudo abrir %s: %v\n", targetPort, err)
			fmt.Println("[REINTENTO] Reintentando en 2 segundos...")
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("[OK] Conectado exitosamente a %s!\n", targetPort)
		fmt.Println("[STREAM] Transmitiendo telemetría continua...")

		// Pausa de estabilización por si el ESP32 se reinició al abrir el puerto
		time.Sleep(1500 * time.Millisecond)

		// Bucle de transmisión
		for {
			payload := collector.GetPayload()
			data, err := json.Marshal(payload)
			if err != nil {
				time.Sleep(interval)
				continue
			}

			// Enviar con salto de línea (\n)
			line := append(data, '\n')
			_, err = port.Write(line)
			if err != nil {
				fmt.Printf("[DESCONECTADO] Error enviando datos a %s: %v\n", targetPort, err)
				port.Close()
				break
			}

			// Lectura no bloqueante de respuestas del ESP32 (debug)
			buf := make([]byte, 256)
			port.SetReadTimeout(50 * time.Millisecond)
			n, _ := port.Read(buf)
			if n > 0 {
				fmt.Printf("[ESP32 RX] %s", string(buf[:n]))
			}

			time.Sleep(interval)
		}

		time.Sleep(2 * time.Second)
	}
}
