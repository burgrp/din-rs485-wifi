# BleRiot RS485 Node Firmware

BleRiot nodes that read a single RS485 device and publish its values as BleRiot
wire registers. There is one firmware per device; they share the RS485 transport
(`uartline`), the node runtime (`rig`) and the device-type metadata (`spec`).

## Firmwares

- `meter/` — Sinotimer 3-phase energy meter over Modbus RTU (`rs485`).
- `bms/` — JK BMS over the JK "NW" protocol (`nwbms`).

Each firmware is a single `package main` split by build tags:

- `//go:build tinygo` — node firmware (`main.go`, `registers.go`).
- `//go:build !tinygo` — host provisioning binary (`provision.go`), which builds
  the hub device type from the compiled register table and runs the BleRiot
  provisioning CLI.

## Shared packages

- `uartline/` — half-duplex RS485 transport (DE/RE toggle, idle-gap framing).
- `rig/` — node runtime: pin map, register store, poll loop, BleRiot bring-up.
- `spec/` — device-type metadata and the on-wire node config.
- `rs485/` — Modbus RTU client (used by `meter`).
- `nwbms/` — JK BMS "NW" protocol frame builder and status decoder (used by `bms`).

## Build

Select the firmware with `CMD` (defaults to `meter`):

```sh
make build CMD=meter     # TinyGo firmware image (meter.elf)
make build CMD=bms       # TinyGo firmware image (bms.elf)
make flash CMD=bms       # program + RTT
make provision CMD=bms   # device onboarding
make new CMD=bms
make hub CMD=bms         # host hub with RF diagnostics
make test                # host-side unit tests
```

## Target

Built for the PY32F003 (`py32f003_32k_4k` / `py32f003x6`). See the `TARGET_*`
variables in the [Makefile](Makefile) to retarget.

## Notes

- Firmware runs under `--gc leaking`: never allocate on the poll path. All hot
  buffers are fixed-size fields allocated once at startup.
- Modbus floats are two 16-bit words in low-word/high-word order.
- Wire tags are permanent once deployed; keep them stable.
