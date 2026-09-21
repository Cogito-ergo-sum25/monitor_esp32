// =========================================================================
// Configuración para TFT_eSPI en Arduino IDE (ILI9341 240x320 en ESP32)
// Si usas PlatformIO, este archivo no es necesario porque los pines ya
// están definidos en platformio.ini como build_flags.
// =========================================================================

#define ILI9341_DRIVER

#define TFT_WIDTH  240
#define TFT_HEIGHT 320

// Pines SPI para ESP32
#define TFT_MISO -1
#define TFT_MOSI 23
#define TFT_SCLK 18
#define TFT_CS   15
#define TFT_DC    2
#define TFT_RST   4

// Fuentes cargadas
#define LOAD_GLCD
#define LOAD_FONT2
#define LOAD_FONT4
#define LOAD_FONT6
#define LOAD_FONT7
#define LOAD_GFXFF
#define SMOOTH_FONT

// Frecuencias SPI
#define SPI_FREQUENCY       40000000
#define SPI_READ_FREQUENCY  20000000

