"""
telemetry.py - Módulo para recolectar métricas de CPU, RAM y GPU en Linux.
Optimizado para sistemas AMD (Ryzen k10temp y Radeon RX sysfs) y extensible.
"""

import glob
import os
import psutil


class TelemetryCollector:
    def __init__(self):
        # Inicializar cálculo de CPU percent para evitar 0.0 en la primera llamada
        psutil.cpu_percent(interval=None)

        # Buscar ruta sysfs de GPU AMD
        self.amd_gpu_card_path = self._find_amd_gpu_path()

    def _find_amd_gpu_path(self):
        """Busca el dispositivo DRM principal con soporte de métricas sysfs."""
        card_dirs = sorted(glob.glob("/sys/class/drm/card[0-9]/device"))
        for card_dir in card_dirs:
            if os.path.exists(os.path.join(card_dir, "gpu_busy_percent")):
                return card_dir
        return None

    def get_cpu_temp(self) -> float:
        """Obtiene la temperatura de la CPU buscando sensores k10temp, coretemp, etc."""
        try:
            temps = psutil.sensors_temperatures()
            if not temps:
                return 0.0

            # Prioridad 1: k10temp (AMD Ryzen)
            if "k10temp" in temps:
                for sensor in temps["k10temp"]:
                    if sensor.label in ("Tctl", "Tdie") or not sensor.label:
                        return round(sensor.current, 1)

            # Prioridad 2: coretemp (Intel)
            if "coretemp" in temps:
                for sensor in temps["coretemp"]:
                    if "Package id 0" in sensor.label or sensor.label == "":
                        return round(sensor.current, 1)
                if temps["coretemp"]:
                    return round(temps["coretemp"][0].current, 1)

            # Prioridad 3: cpu_thermal o acpitz
            for key in ("cpu_thermal", "acpitz", "zenpower"):
                if key in temps and temps[key]:
                    return round(temps[key][0].current, 1)

        except Exception:
            pass
        return 0.0

    def get_cpu_metrics(self) -> dict:
        usage = psutil.cpu_percent(interval=None)
        temp = self.get_cpu_temp()
        return {
            "usage": round(usage, 1),
            "temp": temp
        }

    def get_ram_metrics(self) -> dict:
        mem = psutil.virtual_memory()
        used_gb = mem.used / (1024 ** 3)
        total_gb = mem.total / (1024 ** 3)
        return {
            "usage": round(mem.percent, 1),
            "used": round(used_gb, 1),
            "total": round(total_gb, 1)
        }

    def get_gpu_metrics(self) -> dict:
        usage = 0.0
        temp = 0.0
        vram_used = 0.0
        vram_total = 0.0

        # 1. GPU AMD vía sysfs
        if self.amd_gpu_card_path:
            # Uso de GPU (%)
            busy_file = os.path.join(self.amd_gpu_card_path, "gpu_busy_percent")
            if os.path.exists(busy_file):
                try:
                    with open(busy_file, "r") as f:
                        usage = float(f.read().strip())
                except Exception:
                    pass

            # VRAM
            vram_used_file = os.path.join(self.amd_gpu_card_path, "mem_info_vram_used")
            vram_total_file = os.path.join(self.amd_gpu_card_path, "mem_info_vram_total")
            try:
                if os.path.exists(vram_used_file):
                    with open(vram_used_file, "r") as f:
                        vram_used = float(f.read().strip()) / (1024 ** 3)
                if os.path.exists(vram_total_file):
                    with open(vram_total_file, "r") as f:
                        vram_total = float(f.read().strip()) / (1024 ** 3)
            except Exception:
                pass

        # 2. Temperatura de GPU vía psutil sensors
        try:
            temps = psutil.sensors_temperatures()
            if "amdgpu" in temps and temps["amdgpu"]:
                for s in temps["amdgpu"]:
                    if s.label in ("edge", "junction") or not s.label:
                        temp = round(s.current, 1)
                        break
        except Exception:
            pass

        return {
            "usage": round(usage, 1),
            "temp": temp,
            "vram_used": round(vram_used, 1),
            "vram_total": round(vram_total, 1)
        }

    def get_payload(self) -> dict:
        return {
            "cpu": self.get_cpu_metrics(),
            "gpu": self.get_gpu_metrics(),
            "ram": self.get_ram_metrics()
        }

