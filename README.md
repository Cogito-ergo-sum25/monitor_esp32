# ESP32 Desk Dashboard (PC Monitor + Spotify Widget)

Desk Dashboard para escritorio utilizando un **ESP32 WROOM-32** conectado por SPI a una pantalla TFT LCD de 2.8″ con controlador **ILI9341** (resolución 320×240 en orientación horizontal).

---

## 📌 Hardware y Conexión de Pines (SPI)

| Pin Pantalla TFT (ILI9341) | Pin ESP32 (WROOM-32) | Función / Notas |
| :--- | :--- | :--- |
| **VCC** | **5V / VIN** | Alimentación principal de la pantalla |
| **GND** | **GND** | Tierra común |
| **CS** | **GPIO 15** | Chip Select (SPI) |
| **RESET** | **GPIO 4** | Reset |
| **DC / RS** | **GPIO 2** | Data / Command |
| **SDI (MOSI)** | **GPIO 23** | SPI Master Out Slave In |
| **SCK (CLK)** | **GPIO 18** | SPI Clock |
| **LED** | **3.3V** | Retroiluminación (Backlight) |
| **SDO (MISO)** | *No conectado* | No requerido (-1) |

---

## 📂 Estructura del Proyecto

```text
monitor_esp32/
├── firmware/
│   ├── platformio.ini           # Configuración de compilación con flags TFT_eSPI
│   ├── User_Setup_ILI9341.h     # Configuración para TFT_eSPI en Arduino IDE
│   ├── include/
│   │   └── config.h             # Pines, colores RGB565 y dimensiones de UI
│   └── src/
│       ├── display_ui.h         # Modelos de datos y clase DisplayUI
│       ├── display_ui.cpp       # Renderizado flicker-free con TFT_eSprite
│       └── main.cpp             # Lógica principal, recepción Serial JSON y loop
├── host_service/
│   ├── requirements.txt         # psutil, pyserial
│   ├── telemetry.py             # Recolección CPU (Ryzen), GPU (Radeon RX) y RAM
│   └── main.py                  # Detección de puerto serie y transmisión periódica
├── .gitignore
└── README.md
```

---

## 🚀 Cómo Compilar el Firmware

### Opción A: Con PlatformIO (Recomendado)
No requiere editar archivos internos de librerías. Los pines y parámetros de `TFT_eSPI` ya están inyectados en `platformio.ini`.

1. Abre la carpeta `firmware/` en VS Code con la extensión PlatformIO (o vía CLI `pio run`).
2. Conecta el ESP32 por USB y ejecuta:
   ```bash
   pio run -t upload
   ```

### Opción B: Con Arduino IDE
1. Instala las librerías:
   - **TFT_eSPI** de Bodmer.
   - **ArduinoJson** (v7.x) de Benoit Blanchon.
2. Copia el contenido de `firmware/User_Setup_ILI9341.h` dentro del archivo `User_Setup.h` de tu librería TFT_eSPI (ubicada habitualmente en `~/Arduino/libraries/TFT_eSPI/User_Setup.h`).
3. Abre los archivos en `firmware/src/` y súbelos seleccionando la placa `ESP32 Dev Module`.

---

## 🖥️ Cómo Ejecutar el Servicio Host (Go - Recomendado)

El servicio host está desarrollado en **Go (Golang)** para garantizar la máxima portabilidad: genera un **único archivo ejecutable sin dependencias externas** (no requiere instalar Python, ni pip, ni librerías adicionales).

### 1. Ejecución Directa (Modo Prueba / Simulación)
```bash
cd host_service
go run . --mock
```

### 2. Ejecución Normal (Detecta ESP32 automáticamente)
```bash
cd host_service
go run .
# O con argumentos manuales:
go run . -port /dev/ttyUSB0 -baud 115200 -interval 1000
```

### 3. Compilación de Binarios Distribuidos (Windows / Linux / macOS)
Con el `Makefile` incluido puedes generar ejecutables para cualquier sistema operativo:

```bash
cd host_service

# Compilar para tu máquina Linux actual:
make build-linux      # Genera bin/desk-monitor-linux

# Compilar el ejecutable para usuarios de Windows (.exe):
make build-windows    # Genera bin/desk-monitor-windows.exe

# Compilar para macOS (Apple Silicon):
make build-mac        # Genera bin/desk-monitor-darwin-arm64
```

> **Nota:** También se conserva el prototipo en Python en `host_service/python/` para quien prefiera usar scripts en Python.

El servicio enviará cada 1 segundo un payload JSON como:
```json
{
  "cpu": {"usage": 14.5, "temp": 48.0},
  "gpu": {"usage": 22.0, "temp": 50.0, "vram_used": 2.8, "vram_total": 16.0},
  "ram": {"usage": 30.1, "used": 9.4, "total": 31.2}
}
```
Si el ESP32 se desconecta o reinicia, el servicio reintenta la conexión automáticamente.
