#pragma once

#include <Arduino.h>

// ==========================================
// Configuración de Pantalla y Serial
// ==========================================
#define SCREEN_ROTATION 1      // 1 o 3 para orientación horizontal (320x240)
#define SCREEN_WIDTH    320
#define SCREEN_HEIGHT   240
#define SERIAL_BAUD_RATE 115200

// Timeout para considerar que la PC está desconectada (en milisegundos)
#define TELEMETRY_TIMEOUT_MS 3500

// ==========================================
// Paleta de Colores (RGB565)
// ==========================================
#define COLOR_BG            0x0862  // #0d1117 - Fondo principal oscuro
#define COLOR_CARD_BG       0x18C3  // #161b22 - Fondo de tarjetas
#define COLOR_CARD_BORDER   0x3186  // #30363d - Borde de tarjeta
#define COLOR_TEXT_MUTED    0x8410  // Gris secundario
#define COLOR_TEXT_WHITE    0xFFFF  // Blanco primario

// Colores de acento
#define COLOR_CPU_ACCENT    0x067F  // Cyan / Celeste eléctrico
#define COLOR_GPU_ACCENT    0xF9A6  // Naranja / Coral eléctrico
#define COLOR_RAM_ACCENT    0x3E18  // Verde esmeralda claro
#define COLOR_SPOTIFY_GREEN 0x1DCB  // Verde Spotify oficial (#1db954)
#define COLOR_CLOCK_ACCENT  0xFD20  // Ámbar / Dorado eléctrico (#ffa500)
#define COLOR_ALERT_RED     0xF800  // Rojo para sobrecalentamiento (>85°C)

// ==========================================
// Dimensiones del Layout
// ==========================================
#define CARD_TOP_Y         6
#define CARD_HEIGHT        116
#define CARD_WIDTH          98
#define CARD_SPACING         6

#define CARD_CPU_X         6
#define CARD_GPU_X         (CARD_CPU_X + CARD_WIDTH + CARD_SPACING)   // 110
#define CARD_RAM_X         (CARD_GPU_X + CARD_WIDTH + CARD_SPACING)   // 214

#define SPOTIFY_X          6
#define SPOTIFY_Y          128
#define SPOTIFY_WIDTH      308
#define SPOTIFY_HEIGHT     106
// Fila Inferior: Media Player (Izquierda) + Reloj/Fecha (Derecha)
#define BOTTOM_Y           128
#define BOTTOM_HEIGHT      106

#define MEDIA_X            6
#define MEDIA_Y            BOTTOM_Y
#define MEDIA_WIDTH        202
#define MEDIA_HEIGHT       BOTTOM_HEIGHT

#define CLOCK_X            CARD_RAM_X                                  // 214
#define CLOCK_Y            BOTTOM_Y
#define CLOCK_WIDTH        CARD_WIDTH                                 // 98
#define CLOCK_HEIGHT       BOTTOM_HEIGHT

