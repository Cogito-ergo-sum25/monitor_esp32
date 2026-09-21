#include "display_ui.h"

DisplayUI::DisplayUI(TFT_eSPI &tft_ref)
    : tft(tft_ref),
      sprCpu(&tft_ref),
      sprGpu(&tft_ref),
      sprRam(&tft_ref),
      sprSpotify(&tft_ref) {}

void DisplayUI::begin() {
    tft.init();
    tft.setRotation(SCREEN_ROTATION);
    tft.fillScreen(COLOR_BG);

    // Inicializar Sprites individuales para flicker-free sin saturar RAM
    sprCpu.setColorDepth(16);
    sprCpu.createSprite(CARD_WIDTH, CARD_HEIGHT);

    sprGpu.setColorDepth(16);
    sprGpu.createSprite(CARD_WIDTH, CARD_HEIGHT);

    sprRam.setColorDepth(16);
    sprRam.createSprite(CARD_WIDTH, CARD_HEIGHT);

    sprSpotify.setColorDepth(16);
    sprSpotify.createSprite(SPOTIFY_WIDTH, SPOTIFY_HEIGHT);

    // Dibujar pantalla inicial de bienvenida / standby
    TelemetryData initialTelemetry;
    SpotifyData initialSpotify;
    renderAll(initialTelemetry, initialSpotify);
}

uint16_t DisplayUI::getTempColor(float temp) {
    if (temp >= 85.0f) return COLOR_ALERT_RED;
    if (temp >= 75.0f) return 0xFD20; // Naranja fuerte
    if (temp >= 65.0f) return 0xFFE0; // Amarillo
    return COLOR_TEXT_WHITE;
}

void DisplayUI::drawProgressBar(TFT_eSprite &spr, int x, int y, int w, int h, float percent, uint16_t color) {
    if (percent < 0.0f) percent = 0.0f;
    if (percent > 100.0f) percent = 100.0f;

    // Fondo barra
    spr.fillRoundRect(x, y, w, h, 2, COLOR_CARD_BORDER);
    // Relleno
    int fillW = (int)((w * percent) / 100.0f);
    if (fillW > 2) {
        spr.fillRoundRect(x, y, fillW, h, 2, color);
    }
}

void DisplayUI::formatTime(int ms, char *buffer, size_t bufSize) {
    int total_seconds = ms / 1000;
    int minutes = total_seconds / 60;
    int seconds = total_seconds % 60;
    snprintf(buffer, bufSize, "%02d:%02d", minutes, seconds);
}

void DisplayUI::updateTelemetry(const TelemetryData &data) {
    // --- TARJETA CPU ---
    sprCpu.fillScreen(COLOR_BG);
    sprCpu.fillRoundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 6, COLOR_CARD_BG);
    sprCpu.drawRoundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 6, COLOR_CARD_BORDER);

    sprCpu.setTextDatum(TL_DATUM);
    sprCpu.setTextColor(COLOR_CPU_ACCENT, COLOR_CARD_BG);
    sprCpu.drawString("CPU", 8, 8, 2);

    // Indicador conexión
    sprCpu.fillCircle(CARD_WIDTH - 12, 14, 3, data.connected ? COLOR_CPU_ACCENT : COLOR_TEXT_MUTED);

    sprCpu.setTextDatum(TC_DATUM);
    if (data.connected) {
        sprCpu.setTextColor(COLOR_TEXT_WHITE, COLOR_CARD_BG);
        sprCpu.drawString(String((int)data.cpu_usage) + "%", CARD_WIDTH / 2, 28, 4);
        drawProgressBar(sprCpu, 8, 58, CARD_WIDTH - 16, 5, data.cpu_usage, COLOR_CPU_ACCENT);

        // Temperatura
        sprCpu.setTextColor(getTempColor(data.cpu_temp), COLOR_CARD_BG);
        sprCpu.drawString(String((int)data.cpu_temp) + " `C", CARD_WIDTH / 2, 70, 2);

        sprCpu.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        sprCpu.drawString("Ryzen AMD", CARD_WIDTH / 2, 94, 1);
    } else {
        sprCpu.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        sprCpu.drawString("--%", CARD_WIDTH / 2, 28, 4);
        drawProgressBar(sprCpu, 8, 58, CARD_WIDTH - 16, 5, 0, COLOR_CARD_BORDER);
        sprCpu.drawString("OFFLINE", CARD_WIDTH / 2, 74, 2);
    }
    sprCpu.pushSprite(CARD_CPU_X, CARD_TOP_Y);

    // --- TARJETA GPU ---
    sprGpu.fillScreen(COLOR_BG);
    sprGpu.fillRoundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 6, COLOR_CARD_BG);
    sprGpu.drawRoundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 6, COLOR_CARD_BORDER);

    sprGpu.setTextDatum(TL_DATUM);
    sprGpu.setTextColor(COLOR_GPU_ACCENT, COLOR_CARD_BG);
    sprGpu.drawString("GPU", 8, 8, 2);

    sprGpu.fillCircle(CARD_WIDTH - 12, 14, 3, data.connected ? COLOR_GPU_ACCENT : COLOR_TEXT_MUTED);

    sprGpu.setTextDatum(TC_DATUM);
    if (data.connected) {
        sprGpu.setTextColor(COLOR_TEXT_WHITE, COLOR_CARD_BG);
        sprGpu.drawString(String((int)data.gpu_usage) + "%", CARD_WIDTH / 2, 28, 4);
        drawProgressBar(sprGpu, 8, 58, CARD_WIDTH - 16, 5, data.gpu_usage, COLOR_GPU_ACCENT);

        sprGpu.setTextColor(getTempColor(data.gpu_temp), COLOR_CARD_BG);
        sprGpu.drawString(String((int)data.gpu_temp) + " `C", CARD_WIDTH / 2, 70, 2);

        sprGpu.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        if (data.vram_total_gb > 0) {
            sprGpu.drawString(String(data.vram_used_gb, 1) + "/" + String((int)data.vram_total_gb) + "G VRAM", CARD_WIDTH / 2, 94, 1);
        } else {
            sprGpu.drawString("Radeon RX", CARD_WIDTH / 2, 94, 1);
        }
    } else {
        sprGpu.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        sprGpu.drawString("--%", CARD_WIDTH / 2, 28, 4);
        drawProgressBar(sprGpu, 8, 58, CARD_WIDTH - 16, 5, 0, COLOR_CARD_BORDER);
        sprGpu.drawString("OFFLINE", CARD_WIDTH / 2, 74, 2);
    }
    sprGpu.pushSprite(CARD_GPU_X, CARD_TOP_Y);

    // --- TARJETA RAM ---
    sprRam.fillScreen(COLOR_BG);
    sprRam.fillRoundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 6, COLOR_CARD_BG);
    sprRam.drawRoundRect(0, 0, CARD_WIDTH, CARD_HEIGHT, 6, COLOR_CARD_BORDER);

    sprRam.setTextDatum(TL_DATUM);
    sprRam.setTextColor(COLOR_RAM_ACCENT, COLOR_CARD_BG);
    sprRam.drawString("RAM", 8, 8, 2);

    sprRam.fillCircle(CARD_WIDTH - 12, 14, 3, data.connected ? COLOR_RAM_ACCENT : COLOR_TEXT_MUTED);

    sprRam.setTextDatum(TC_DATUM);
    if (data.connected) {
        sprRam.setTextColor(COLOR_TEXT_WHITE, COLOR_CARD_BG);
        sprRam.drawString(String((int)data.ram_usage) + "%", CARD_WIDTH / 2, 28, 4);
        drawProgressBar(sprRam, 8, 58, CARD_WIDTH - 16, 5, data.ram_usage, COLOR_RAM_ACCENT);

        sprRam.setTextColor(COLOR_TEXT_WHITE, COLOR_CARD_BG);
        sprRam.drawString(String(data.ram_used_gb, 1) + " GB", CARD_WIDTH / 2, 70, 2);

        sprRam.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        sprRam.drawString("de " + String((int)data.ram_total_gb) + " GB", CARD_WIDTH / 2, 94, 1);
    } else {
        sprRam.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        sprRam.drawString("--%", CARD_WIDTH / 2, 28, 4);
        drawProgressBar(sprRam, 8, 58, CARD_WIDTH - 16, 5, 0, COLOR_CARD_BORDER);
        sprRam.drawString("OFFLINE", CARD_WIDTH / 2, 74, 2);
    }
    sprRam.pushSprite(CARD_RAM_X, CARD_TOP_Y);
}

void DisplayUI::updateSpotify(const SpotifyData &spotify) {
    sprSpotify.fillScreen(COLOR_BG);
    sprSpotify.fillRoundRect(0, 0, SPOTIFY_WIDTH, SPOTIFY_HEIGHT, 6, COLOR_CARD_BG);
    sprSpotify.drawRoundRect(0, 0, SPOTIFY_WIDTH, SPOTIFY_HEIGHT, 6, COLOR_CARD_BORDER);

    // Cabecera Spotify
    sprSpotify.fillCircle(14, 15, 4, COLOR_SPOTIFY_GREEN);
    sprSpotify.setTextDatum(TL_DATUM);
    sprSpotify.setTextColor(COLOR_SPOTIFY_GREEN, COLOR_CARD_BG);
    sprSpotify.drawString("SPOTIFY", 24, 8, 2);

    // Badge de estado a la derecha
    sprSpotify.setTextDatum(TR_DATUM);
    if (spotify.is_playing) {
        sprSpotify.setTextColor(COLOR_SPOTIFY_GREEN, COLOR_CARD_BG);
        sprSpotify.drawString("PLAYING", SPOTIFY_WIDTH - 12, 8, 2);
    } else {
        sprSpotify.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
        sprSpotify.drawString("PAUSED", SPOTIFY_WIDTH - 12, 8, 2);
    }

    // Título de la canción
    sprSpotify.setTextDatum(TL_DATUM);
    sprSpotify.setTextColor(COLOR_TEXT_WHITE, COLOR_CARD_BG);
    String title = spotify.title;
    if (title.length() > 25) {
        title = title.substring(0, 22) + "...";
    }
    sprSpotify.drawString(title, 12, 30, 4);

    // Nombre del artista
    sprSpotify.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
    String artist = spotify.artist;
    if (artist.length() > 34) {
        artist = artist.substring(0, 31) + "...";
    }
    sprSpotify.drawString(artist, 12, 58, 2);

    // Barra de reproducción y tiempos
    char curTime[10], totTime[10];
    formatTime(spotify.progress_ms, curTime, sizeof(curTime));
    formatTime(spotify.duration_ms, totTime, sizeof(totTime));

    sprSpotify.setTextDatum(TL_DATUM);
    sprSpotify.setTextColor(COLOR_TEXT_MUTED, COLOR_CARD_BG);
    sprSpotify.drawString(curTime, 12, 82, 1);

    int barX = 50;
    int barW = SPOTIFY_WIDTH - 100;
    float progressPercent = (spotify.duration_ms > 0)
                                ? ((float)spotify.progress_ms / (float)spotify.duration_ms) * 100.0f
                                : 0.0f;
    drawProgressBar(sprSpotify, barX, 84, barW, 5, progressPercent, COLOR_SPOTIFY_GREEN);

    sprSpotify.setTextDatum(TR_DATUM);
    sprSpotify.drawString(totTime, SPOTIFY_WIDTH - 12, 82, 1);

    sprSpotify.pushSprite(SPOTIFY_X, SPOTIFY_Y);
}

void DisplayUI::renderAll(const TelemetryData &telemetry, const SpotifyData &spotify) {
    updateTelemetry(telemetry);
    updateSpotify(spotify);
}

