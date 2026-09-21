// ESP32 Desk Dashboard - Client Application

let txCount = 0;
let eventSource = null;

document.addEventListener('DOMContentLoaded', () => {
  initSSE();
  fetchPorts();
  setupEventListeners();
});

// Inicialización de Server-Sent Events (SSE) para streaming continuo de telemetría
function initSSE() {
  if (eventSource) {
    eventSource.close();
  }

  eventSource = new EventSource('/api/events');

  eventSource.onopen = () => {
    updateConnectionStatus(true, 'Conectado al Host en Go');
  };

  eventSource.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data);
      handleTelemetryUpdate(data);
    } catch (err) {
      console.error('Error parseando SSE telemetry:', err);
    }
  };

  eventSource.onerror = () => {
    updateConnectionStatus(false, 'Reconectando con el host...');
    // Fallback de polling mientras se restablece SSE
    setTimeout(fetchStatus, 3000);
  };
}

// Actualización reactiva de la interfaz web y del espejo virtual de la pantalla
function handleTelemetryUpdate(payload) {
  txCount++;
  document.getElementById('stat-tx-count').textContent = txCount;

  const { cpu, gpu, ram, device } = payload;

  // 1. Espejo Virtual del TFT (320x240)
  if (cpu) {
    const cpuVal = Math.round(cpu.usage);
    document.getElementById('tft-cpu-usage').textContent = `${cpuVal}%`;
    document.getElementById('tft-cpu-bar').style.width = `${Math.min(100, Math.max(0, cpuVal))}%`;
    document.getElementById('tft-cpu-temp').textContent = `${Math.round(cpu.temp)} °C`;
    document.getElementById('tft-dot-cpu').style.background = 'var(--accent-cpu)';

    // Extended Stats
    document.getElementById('ext-cpu-usage').textContent = `${cpuVal}%`;
    document.getElementById('ext-cpu-bar').style.width = `${cpuVal}%`;
    document.getElementById('ext-cpu-temp').textContent = `${cpu.temp.toFixed(1)}°C`;
  }

  if (gpu) {
    const gpuVal = Math.round(gpu.usage);
    document.getElementById('tft-gpu-usage').textContent = `${gpuVal}%`;
    document.getElementById('tft-gpu-bar').style.width = `${Math.min(100, Math.max(0, gpuVal))}%`;
    document.getElementById('tft-gpu-temp').textContent = `${Math.round(gpu.temp)} °C`;
    document.getElementById('tft-dot-gpu').style.background = 'var(--accent-gpu)';
    if (gpu.vram_total > 0) {
      document.getElementById('tft-vram-text').textContent = `${gpu.vram_used.toFixed(1)}/${Math.round(gpu.vram_total)}G VRAM`;
      document.getElementById('ext-gpu-vram').textContent = `VRAM: ${gpu.vram_used.toFixed(1)} / ${gpu.vram_total.toFixed(1)} GB`;
    }

    // Extended Stats
    document.getElementById('ext-gpu-usage').textContent = `${gpuVal}%`;
    document.getElementById('ext-gpu-bar').style.width = `${gpuVal}%`;
    document.getElementById('ext-gpu-temp').textContent = `${gpu.temp.toFixed(1)}°C`;
  }

  if (ram) {
    const ramVal = Math.round(ram.usage);
    document.getElementById('tft-ram-usage').textContent = `${ramVal}%`;
    document.getElementById('tft-ram-bar').style.width = `${Math.min(100, Math.max(0, ramVal))}%`;
    document.getElementById('tft-ram-used').textContent = `${ram.used.toFixed(1)} GB`;
    document.getElementById('tft-ram-total').textContent = `de ${Math.round(ram.total)} GB`;
    document.getElementById('tft-dot-ram').style.background = 'var(--accent-ram)';

    // Extended Stats
    document.getElementById('ext-ram-usage').textContent = `${ramVal}%`;
    document.getElementById('ext-ram-bar').style.width = `${ramVal}%`;
    document.getElementById('ext-ram-used').textContent = `${ram.used.toFixed(1)} GB`;
    document.getElementById('ext-ram-total').textContent = `Total: ${ram.total.toFixed(1)} GB`;
  }

  // Media (Spotify / Player)
  if (media && media.title) {
    document.getElementById('tft-song-title').textContent = media.title;
    document.getElementById('tft-song-artist').textContent = media.artist || 'Artista';
    document.getElementById('tft-spotify-status').textContent = media.is_playing ? 'PLAYING' : 'PAUSED';
    document.getElementById('tft-spotify-status').style.color = media.is_playing ? 'var(--accent-spotify)' : 'var(--text-muted)';

    if (media.duration_ms > 0) {
      const curMin = Math.floor(media.progress_ms / 60000);
      const curSec = Math.floor((media.progress_ms % 60000) / 1000);
      const totMin = Math.floor(media.duration_ms / 60000);
      const totSec = Math.floor((media.duration_ms % 60000) / 1000);

      document.getElementById('tft-cur-time').textContent = 
        `${curMin}:${curSec < 10 ? '0' : ''}${curSec}`;
      document.getElementById('tft-tot-time').textContent = 
        `${totMin}:${totSec < 10 ? '0' : ''}${totSec}`;

      const pct = (media.progress_ms / media.duration_ms) * 100;
      document.getElementById('tft-song-bar').style.width = `${Math.min(100, Math.max(0, pct))}%`;
    }
  }

  // Estado del Dispositivo Físico
  if (device) {
    if (device.connected) {
      updateConnectionStatus(true, `ESP32 Activo (${device.port})`);
      document.getElementById('current-port').textContent = device.port || 'Auto';
    } else {
      updateConnectionStatus(false, 'Esperando conexión de ESP32...');
      document.getElementById('current-port').textContent = 'Desconectado';
    }
  }
}

function updateConnectionStatus(isConnected, text) {
  const dot = document.getElementById('status-dot');
  const statusText = document.getElementById('status-text');

  if (isConnected) {
    dot.classList.add('connected');
    statusText.textContent = text;
  } else {
    dot.classList.remove('connected');
    statusText.textContent = text;
  }
}

// Consultar lista de puertos disponibles
async function fetchPorts() {
  try {
    const res = await fetch('/api/ports');
    const data = await res.json();
    const select = document.getElementById('port-select');

    select.innerHTML = '<option value="">Detección automática (Recomendado)</option>';
    if (data.ports && data.ports.length > 0) {
      data.ports.forEach(p => {
        const opt = document.createElement('option');
        opt.value = p;
        opt.textContent = p;
        select.appendChild(opt);
      });
    }

    if (data.current) {
      select.value = data.current;
    }
  } catch (err) {
    console.error('Error obteniendo puertos:', err);
  }
}

// Consultar estado general vía REST
async function fetchStatus() {
  try {
    const res = await fetch('/api/status');
    const data = await res.json();
    if (data.device) {
      updateConnectionStatus(data.device.connected, data.device.connected ? `ESP32 Activo (${data.device.port})` : 'ESP32 Desconectado');
      document.getElementById('current-port').textContent = data.device.port || 'Auto';
    }
  } catch (err) {
    updateConnectionStatus(false, 'Servicio Host Desconectado');
  }
}

function setupEventListeners() {
  // Botón refrescar puertos
  document.getElementById('btn-refresh-ports').addEventListener('click', () => {
    fetchPorts();
  });

  // Selector de puerto
  document.getElementById('port-select').addEventListener('change', async (e) => {
    const selectedPort = e.target.value;
    try {
      await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ port: selectedPort })
      });
    } catch (err) {
      console.error('Error guardando puerto:', err);
    }
  });

  // Botón Spotify (placeholder)
  document.getElementById('btn-spotify-connect').addEventListener('click', () => {
    alert('Próxima Fase: El flujo OAuth de Spotify permitirá enlazar tu cuenta para sincronizar automáticamente títulos, portadas y reproductor!');
  });
}

