#pragma once

#include <Arduino.h>
#include <TFT_eSPI.h>
#include "config.h"

struct TelemetryData {
    // CPU
    float cpu_usage = 0.0f;
    float cpu_temp = 0.0f;

    // GPU
    float gpu_usage = 0.0f;
    float gpu_temp = 0.0f;
    float vram_used_gb = 0.0f;
    float vram_total_gb = 0.0f;

    // RAM
    float ram_usage = 0.0f;
    float ram_used_gb = 0.0f;
    float ram_total_gb = 0.0f;

    // Conexión
    bool connected = false;
    unsigned long last_packet_time = 0;
};

struct SpotifyData {
    String title = "Desk Dashboard";
    String artist = "Spotify en espera (Wi-Fi)";
    int progress_ms = 0;
    int duration_ms = 1000;
    bool is_playing = false;
    bool wifi_connected = false;
};

class DisplayUI {
public:
    DisplayUI(TFT_eSPI &tft);
    void begin();
    void updateTelemetry(const TelemetryData &data);
    void updateSpotify(const SpotifyData &data);
    void renderAll(const TelemetryData &telemetry, const SpotifyData &spotify);

private:
    TFT_eSPI &tft;
    TFT_eSprite sprCpu;
    TFT_eSprite sprGpu;
    TFT_eSprite sprRam;
    TFT_eSprite sprSpotify;

    void drawProgressBar(TFT_eSprite &spr, int x, int y, int w, int h, float percent, uint16_t color);
    void formatTime(int ms, char *buffer, size_t bufSize);
    uint16_t getTempColor(float temp);
};

