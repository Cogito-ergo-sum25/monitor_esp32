#include <Arduino.h>
#include <TFT_eSPI.h>
#include <ArduinoJson.h>
#include "config.h"
#include "display_ui.h"

TFT_eSPI tft = TFT_eSPI();
DisplayUI ui(tft);

TelemetryData telemetry;
SpotifyData spotify;

unsigned long lastSpotifyTick = 0;
unsigned long lastStatusCheck = 0;

void setup() {
    Serial.begin(SERIAL_BAUD_RATE);
    Serial.setTimeout(50);
    delay(200);

    Serial.println();
    Serial.println("=========================================");
    Serial.println("  ESP32 Desk Dashboard Initializing...   ");
    Serial.println("=========================================");

    // Inicializar pantalla y sprites
    ui.begin();

    // Valores iniciales de Spotify (placeholder visual hasta fase Wi-Fi)
    spotify.title = "Starboy";
    spotify.artist = "The Weeknd, Daft Punk";
    spotify.duration_ms = 230000;
    spotify.progress_ms = 45000;
    spotify.is_playing = true;
    spotify.wifi_connected = false;

    ui.renderAll(telemetry, spotify);
    Serial.println("Display UI listo. Esperando telemetria JSON...");
}

void loop() {
    // 1. Lectura del puerto serie para datos de telemetría de la PC
    if (Serial.available() > 0) {
        String jsonLine = Serial.readStringUntil('\n');
        jsonLine.trim();

        if (jsonLine.length() > 0 && jsonLine.startsWith("{")) {
            JsonDocument doc;
            DeserializationError error = deserializeJson(doc, jsonLine);

            if (!error) {
                // Lectura CPU
                telemetry.cpu_usage = doc["cpu"]["usage"] | telemetry.cpu_usage;
                telemetry.cpu_temp  = doc["cpu"]["temp"]  | telemetry.cpu_temp;

                // Lectura GPU
                telemetry.gpu_usage    = doc["gpu"]["usage"]      | telemetry.gpu_usage;
                telemetry.gpu_temp     = doc["gpu"]["temp"]       | telemetry.gpu_temp;
                telemetry.vram_used_gb = doc["gpu"]["vram_used"]  | telemetry.vram_used_gb;
                telemetry.vram_total_gb= doc["gpu"]["vram_total"] | telemetry.vram_total_gb;

                // Lectura RAM
                telemetry.ram_usage    = doc["ram"]["usage"] | telemetry.ram_usage;
                telemetry.ram_used_gb  = doc["ram"]["used"]  | telemetry.ram_used_gb;
                telemetry.ram_total_gb = doc["ram"]["total"] | telemetry.ram_total_gb;

                // Lectura Media (Spotify / MPRIS en vivo)
                if (doc["media"].is<JsonObject>()) {
                    const char* titleStr = doc["media"]["title"];
                    const char* artistStr = doc["media"]["artist"];
                    if (titleStr && strlen(titleStr) > 0) {
                        spotify.title = String(titleStr);
                    }
                    if (artistStr && strlen(artistStr) > 0) {
                        spotify.artist = String(artistStr);
                    }
                    spotify.is_playing = doc["media"]["is_playing"] | spotify.is_playing;
                    spotify.progress_ms = doc["media"]["progress_ms"] | spotify.progress_ms;
                    spotify.duration_ms = doc["media"]["duration_ms"] | spotify.duration_ms;
                    ui.updateSpotify(spotify);
                }

                telemetry.connected = true;
                telemetry.last_packet_time = millis();

                // Actualizar tarjetas de métricas inmediatamente
                ui.updateTelemetry(telemetry);
            } else {
                // Error de JSON parsing opcional para debug
                // Serial.printf("JSON parse error: %s\n", error.c_str());
            }
        }
    }

    // 2. Comprobar timeout de conexión con la PC
    if (telemetry.connected && (millis() - telemetry.last_packet_time > TELEMETRY_TIMEOUT_MS)) {
        telemetry.connected = false;
        ui.updateTelemetry(telemetry);
        Serial.println("[WARN] Telemetria de PC desconectada (timeout)");
    }

    // 3. Avance de barra de Spotify cuando la PC esté desconectada (placeholder)
    if (millis() - lastSpotifyTick >= 1000) {
        lastSpotifyTick = millis();
        if (spotify.is_playing && !telemetry.connected) {
            spotify.progress_ms += 1000;
            if (spotify.progress_ms > spotify.duration_ms) {
                spotify.progress_ms = 0;
            }
            ui.updateSpotify(spotify);
        }
    }

    // Pequeño delay de cortesía para el planificador de tareas de FreeRTOS
    delay(5);
}

