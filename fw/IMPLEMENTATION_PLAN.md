# Implementation Plan: ESP32 MicroPython -> BleRiot/PY32

## 1) Module and package layout

Use one device module mirroring the BleRiot example pattern:

- `fw/spec`: shared device type metadata and register tags.
- `fw`: flat root-level `package main` split by build tags.
  - `main.go` (`tinygo`) for node firmware.
  - `test-hub.go` (`!tinygo`) for host inventory/hub startup.

## 2) RS485->Registry bridge model in BleRiot

- Firmware polls Modbus RTU devices on RS485 periodically.
- Poller decodes register words to float32, scales to int32 for BleRiot wire registers.
- Register store is keyed by permanent wire tag (uint16).
- BleRiot node runtime publishes these values.
- Host `lib/site` hub bridges tag values to Registry provider/consumer.

## 3) Config format suggestion

- Keep inventory as Go code for deployed runtime (BleRiot-native).
- Keep YAML examples for operator-facing source configuration and generation inputs.
- Files in `config/`:
  - `site.example.yaml`: site/radio/registry endpoint defaults.
  - `device.example.yaml`: node pins, uid/key, Modbus devices, register map.

## 4) Practical next integration steps

1. Add `github.com/burgrp/bleriot/lib` dependency.
2. Replace `noopClient` with UART RTU implementation for TinyGo target.
3. Bind `registerStore` to concrete `lib/node` register backing by wire tag.
4. Implement host inventory in `test-hub.go` and call `cli.Start`.
5. Add provisioning workflow (`new` + `provision`) and check radio bring-up.

## 5) Board/runtime constraints

- Keep firmware paths free of `errors.Is`/`errors.As` and interface comparisons.
- Do not construct `time.Time` on firmware path; use duration sleeps and monotonic nanos if needed.
- Preserve tag stability permanently once deployed.
- Keep hub/runtime generic; device-specific logic only in module `spec` and polling decoder.
- Expect tight flash/RAM budget on PY32F003W1xSx; measure size frequently during integration.
