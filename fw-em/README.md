# BleRiot RS485 energy-meter firmware

`fw-em` is the BleRiot node firmware for one three-phase DIN-rail energy meter
connected to one Bleriot RS485 board. The same source builds the independently
provisioned `em.grid` and `em.house` nodes; each image polls only its local meter.

## Registry contract

The two inventory instance names preserve the deployed Registry prefixes:
`em.grid.*` and `em.house.*`. Both expose the same 20 read-only measurements
defined in `spec/device_type.go`, including the existing titles and units.
Wire tags remain the historical Modbus word address plus one.

Values travel over BleRiot as the meter's raw IEEE-754 float32 bits. The hub
converts them to Registry numbers and maps NaN or infinity to `nil`.

## Meter protocol

The implementation follows pages 10-13 of
`../doc/Three phase four wire din-rail energy meter.pdf`:

- Modbus RTU function `0x04` for input registers
- slave address 1, 9600 baud, 8 data bits, even parity, 1 stop bit by default
- two Modbus words per IEEE-754 float32 value
- seven contiguous reads for the 20 measurements

The manual's `43 6B 58 0E` example is high-word-first and is the default.
Set `WordOrderLowFirst` in an instance config only when a meter is verified to
use swapped 16-bit words. After three failed scans by default, affected values
become null until a valid response returns.

## Board wiring

The board schematic connects USART1 TX/RX to `PB6`/`PB7` and the ST3485E
direction input to `PF4`. The PAN211x uses `PF3` for CSN, `PA2` for SCK, and
`PA3` for bidirectional data. J4 is GND, B, A on pins 1, 2, 3.

The board has no RS485 isolation, termination, or fail-safe bias network. Use a
common ground and provide termination/bias appropriate to the bus topology.

The inventory currently selects `PY32F030x8` (64 KiB flash, 8 KiB SRAM).
Confirm the fitted MCU density marking before flashing and change `spec.Chip`
if the assembled part has a different suffix.

## Build and flash

```sh
go test ./...
go run . gen em.grid
go run . make em.grid build
go run . make em.house build
go run . make em.grid flash
```

`make` is driven through the BleRiot CLI so it generates `main_gen.go` with the
selected node's address, key, RF channel, and meter configuration before each
build. Start the bridge with the Registry endpoint used by the site:

```sh
go run . hub --registry http://registry-host:port
```