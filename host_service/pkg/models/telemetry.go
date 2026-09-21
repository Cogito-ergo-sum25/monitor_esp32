package models

// CPUMetrics representa las métricas de uso y temperatura del procesador.
type CPUMetrics struct {
	Usage float64 `json:"usage"`
	Temp  float64 `json:"temp"`
}

// GPUMetrics representa las métricas de la tarjeta gráfica y memoria de video.
type GPUMetrics struct {
	Usage     float64 `json:"usage"`
	Temp      float64 `json:"temp"`
	VramUsed  float64 `json:"vram_used"`
	VramTotal float64 `json:"vram_total"`
}

// RAMMetrics representa el uso de la memoria RAM del sistema en GB y porcentaje.
type RAMMetrics struct {
	Usage float64 `json:"usage"`
	Used  float64 `json:"used"`
	Total float64 `json:"total"`
}

// MediaMetrics representa la información del reproductor multimedia activo (Spotify, navegador, etc.).
type MediaMetrics struct {
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	IsPlaying  bool   `json:"is_playing"`
	ProgressMs int    `json:"progress_ms"`
	DurationMs int    `json:"duration_ms"`
}

// ClockMetrics representa la hora y fecha local de la PC.
type ClockMetrics struct {
	Time string `json:"time"`
	Date string `json:"date"`
	Day  string `json:"day"`
}

// TelemetryPayload es la estructura enviada por serial en formato JSON al ESP32.
type TelemetryPayload struct {
	CPU   CPUMetrics   `json:"cpu"`
	GPU   GPUMetrics   `json:"gpu"`
	RAM   RAMMetrics   `json:"ram"`
	Media MediaMetrics `json:"media"`
	Clock ClockMetrics `json:"clock"`
}
