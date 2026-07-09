# BleRiot RS485 Node Scaffold

This module is a practical migration scaffold from the existing ESP32 MicroPython bridge to a BleRiot node model.

## Build Modes

- `//go:build tinygo` in `main.go` builds the node firmware.
- `//go:build !tinygo` in `test-hub.go` builds the host site binary.
- This mirrors the BleRiot example convention: one flat `package main` split by build tags.

## Expected BleRiot Imports

These are the canonical package paths to integrate once `github.com/burgrp/bleriot` is available in your workspace:

- `github.com/burgrp/bleriot/lib/node`
- `github.com/burgrp/bleriot/lib/node/pan211x`
- `github.com/burgrp/bleriot/lib/shared/config`
- `github.com/burgrp/bleriot/lib/shared/inventory`
- `github.com/burgrp/bleriot/lib/site/cli`

## Migration Model

- Node firmware polls RS485/Modbus devices periodically.
- Poll results are decoded and scaled into integer BleRiot register values, keyed by permanent wire `Tag` (uint16).
- Node runtime publishes these registers over BleRiot radio.
- Host hub bridges generic BleRiot tags to Registry fields; no device-specific hub logic.

## Assumptions

- Exact `lib/node` runtime APIs are not hardcoded in this scaffold where uncertain.
- Pin mapping from KiCad is applied in `board_py32f003.go`.
- Floating Modbus values are interpreted as two 16-bit words in low-word/high-word order, matching legacy Python behavior.

## Quick Start

1. Start with local compile check:
   - `cd fw && go test ./...`
2. Use the BleRiot-style Makefile targets:
   - `make build` (TinyGo firmware image)
   - `make flash` (program + RTT)
   - `make hub` (host hub with diagnostics)
   - `make provision` / `make new` (device onboarding)
