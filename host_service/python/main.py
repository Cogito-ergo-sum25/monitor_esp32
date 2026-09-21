#!/usr/bin/env python3
"""
main.py - Servicio host de telemetría para ESP32 Desk Dashboard.
Detecta automáticamente el puerto serie, recopila métricas y envía JSON a 115200 baud.
"""

import argparse
import json
import sys
import time
import serial
import serial.tools.list_ports

from telemetry import TelemetryCollector


def find_esp32_port():
    """Busca automáticamente un puerto serie conectado a un ESP32 o adaptador USB-UART."""
    ports = list(serial.tools.list_ports.comports())
    if not ports:
        return None

    # Prioridad: Dispositivos conocidos de ESP32 / CH340 / CP210x / FTDI / USB ACM
    keywords = ["cp210", "ch340", "ch341", "ftdi", "esp32", "usb to uart", "uart"]
    for port in ports:
        desc = (port.description or "").lower()
        hwid = (port.hwid or "").lower()
        if any(k in desc or k in hwid for k in keywords):
            return port.device

    # Si no hay coincidencias exactas en la descripción, busca /dev/ttyUSB* o /dev/ttyACM*
    for port in ports:
        if "ttyUSB" in port.device or "ttyACM" in port.device:
            return port.device

    return ports[0].device if ports else None


def run_service(port=None, baud=115200, interval=1.0, mock_serial=False):
    collector = TelemetryCollector()

    print("==================================================")
    print("  ESP32 Desk Dashboard - Host Telemetry Service   ")
    print("==================================================")
    print(f"Intervalo de actualización: {interval}s")
    print(f"Baud rate: {baud}")

    if mock_serial:
        print("[MODO DE PRUEBA] Simulación activa (--mock-serial). No se usará puerto serie.")
        while True:
            payload = collector.get_payload()
            json_str = json.dumps(payload)
            print(f"[JSON TX] {json_str}")
            time.sleep(interval)

    # Bucle continuo con auto-reconexión
    while True:
        target_port = port or find_esp32_port()

        if not target_port:
            print("[INFO] Esperando conexión de ESP32 (/dev/ttyUSB* o /dev/ttyACM*)...")
            time.sleep(2.0)
            continue

        print(f"[CONEXIÓN] Intentando abrir puerto serie: {target_port} a {baud} baud...")
        try:
            with serial.Serial(target_port, baud, timeout=1.0) as ser:
                print(f"[OK] Conectado exitosamente a {target_port}!")
                print("[STREAM] Transmitiendo telemetría...")

                # Pequeña pausa para permitir que el ESP32 se reinicie si DTR/RTS conmutó
                time.sleep(1.5)

                while True:
                    payload = collector.get_payload()
                    line = json.dumps(payload) + "\n"
                    ser.write(line.encode("utf-8"))
                    ser.flush()

                    # Opcional: leer respuestas o logs que mande el ESP32
                    while ser.in_waiting > 0:
                        incoming = ser.readline().decode("utf-8", errors="replace").strip()
                        if incoming:
                            print(f"[ESP32 RX] {incoming}")

                    time.sleep(interval)

        except (serial.SerialException, OSError) as e:
            print(f"[DESCONECTADO] Puerto {target_port} no disponible o desconectado: {e}")
            print("[REINTENTO] Reintentando en 2 segundos...")
            time.sleep(2.0)
        except KeyboardInterrupt:
            print("\n[SALIR] Servicio detenido por el usuario.")
            sys.exit(0)


def main():
    parser = argparse.ArgumentParser(description="Servicio host de telemetría para ESP32 Desk Dashboard.")
    parser.add_argument("-p", "--port", type=str, default=None, help="Puerto serie manual (ej. /dev/ttyUSB0).")
    parser.add_argument("-b", "--baud", type=int, default=115200, help="Velocidad en baudios (default: 115200).")
    parser.add_argument("-i", "--interval", type=float, default=1.0, help="Intervalo en segundos (default: 1.0s).")
    parser.add_argument("--mock-serial", action="store_true", help="Imprime JSON en consola sin enviar a serial.")

    args = parser.parse_args()
    run_service(port=args.port, baud=args.baud, interval=args.interval, mock_serial=args.mock_serial)


if __name__ == "__main__":
    main()

