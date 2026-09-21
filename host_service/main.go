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
	"monitor-esp32-host/pkg/web"
)

func main() {
	portFlag := flag.String("port", "", "Puerto serie manual (ej. /dev/ttyUSB0 en Linux o COM3 en Windows)")
	baudFlag := flag.Int("baud", 115200, "Velocidad en baudios")
	intervalFlag := flag.Int("interval", 1000, "Intervalo de actualización en milisegundos")
	webPortFlag := flag.Int("web-port", 8080, "Puerto para la interfaz web de control (default: 8080)")
	noWebFlag := flag.Bool("no-web", false, "Desactivar la interfaz web local")
	mockFlag := flag.Bool("mock", false, "Modo de simulación (imprime JSON en consola sin usar puerto serie)")
	flag.Parse()

	collector := telemetry.NewCollector()
	hub := web.NewHub(*baudFlag, *intervalFlag)

	fmt.Println("==================================================")
	fmt.Println("   ESP32 Desk Dashboard - Host Service (Go)       ")
	fmt.Println("==================================================")
	fmt.Printf("Intervalo de actualización: %d ms\n", *intervalFlag)
	fmt.Printf("Baud rate: %d\n", *baudFlag)

	// Iniciar servidor web embebido (HTML, CSS, JS, SSE)
	if !*noWebFlag {
		webServer := web.NewServer(*webPortFlag, hub)
		go func() {
			if err := webServer.Start(); err != nil {
				fmt.Printf("[WEB ERROR] No se pudo iniciar el servidor web: %v\n", err)
			}
		}()
	}

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
		hub.UpdateState(true, "MOCK-SIMULATOR")
		for {
			payload := collector.GetPayload()
			hub.Broadcast(payload)

			data, err := json.Marshal(payload)
			if err == nil {
				fmt.Printf("[JSON TX] %s\n", string(data))
			}
			time.Sleep(interval)
		}
	}

	// 2. Modo Real con Auto-Reconexión Continua y Streaming hacia ESP32 y Web
	mode := &serial.Mode{
		BaudRate: *baudFlag,
	}

	activePort := *portFlag

	for {
		// Comprobar si el usuario cambió el puerto desde la interfaz web
		select {
		case newPort := <-hub.PortChangeCh:
			activePort = newPort
			fmt.Printf("[WEB SETTINGS] Cambio de puerto solicitado a: %s\n", activePort)
		default:
		}

		targetPort := activePort
		if targetPort == "" {
			targetPort = serialutil.FindESP32Port()
		}

		if targetPort == "" {
			hub.UpdateState(false, "")
			fmt.Println("[INFO] Esperando conexión de ESP32 (puerto serie no detectado)...")
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("[CONEXIÓN] Intentando abrir puerto: %s a %d baud...\n", targetPort, *baudFlag)
		port, err := serial.Open(targetPort, mode)
		if err != nil {
			hub.UpdateState(false, targetPort)
			fmt.Printf("[ERROR] No se pudo abrir %s: %v\n", targetPort, err)
			fmt.Println("[REINTENTO] Reintentando en 2 segundos...")
			time.Sleep(2 * time.Second)
			continue
		}

		hub.UpdateState(true, targetPort)
		fmt.Printf("[OK] Conectado exitosamente a %s!\n", targetPort)
		fmt.Println("[STREAM] Transmitiendo telemetría continua al ESP32 y Dashboard Web...")

		// Pausa de estabilización por si el ESP32 se reinició al abrir el puerto
		time.Sleep(1500 * time.Millisecond)

		// Goroutine para leer respuestas del ESP32 en segundo plano sin bloquear el loop
		stopReader := make(chan struct{})
		go func() {
			buf := make([]byte, 256)
			for {
				select {
				case <-stopReader:
					return
				default:
					n, err := port.Read(buf)
					if err != nil {
						return
					}
					if n > 0 {
						fmt.Printf("[ESP32 RX] %s", string(buf[:n]))
					}
				}
			}
		}()

		// Bucle de transmisión
		for {
			// Comprobar si se solicitó un cambio de puerto desde la web
			var switchPort string
			select {
			case switchPort = <-hub.PortChangeCh:
			default:
			}
			if switchPort != "" && switchPort != targetPort {
				activePort = switchPort
				close(stopReader)
				port.Close()
				hub.UpdateState(false, targetPort)
				break
			}

			payload := collector.GetPayload()
			// Transmitir a la interfaz web (SSE)
			hub.Broadcast(payload)

			data, err := json.Marshal(payload)
			if err != nil {
				time.Sleep(interval)
				continue
			}

			// Enviar con salto de línea (\n) al ESP32
			line := append(data, '\n')
			_, err = port.Write(line)
			if err != nil {
				fmt.Printf("[DESCONECTADO] Error enviando datos a %s: %v\n", targetPort, err)
				close(stopReader)
				port.Close()
				hub.UpdateState(false, targetPort)
				break
			}

			fmt.Printf("[TX] CPU: %.1f%% (%.1f°C) | GPU: %.0f%% (%.0f°C) | RAM: %.1f%%\n",
				payload.CPU.Usage, payload.CPU.Temp,
				payload.GPU.Usage, payload.GPU.Temp,
				payload.RAM.Usage)

			time.Sleep(interval)
		}

		time.Sleep(2 * time.Second)
	}
}
