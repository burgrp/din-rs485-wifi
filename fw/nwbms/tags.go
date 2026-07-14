// Package nwbms decodes the JK BMS "NW" RS485 protocol (frames starting with
// 0x4E 0x57). It is a straight port of the proven syssi/esphome-jk-bms driver
// (the jk_modbus transport plus the jk_bms read-all status decode). The frame
// building and decoding are board-independent and host-testable; only the
// Client (which drives a uartline.Line) is firmware-only.
package nwbms

// Wire scale dividers used by the emitted integer values. Voltages and currents
// are emitted in milli-units; counts, percentages, temperatures and bitmasks
// are emitted as plain integers.
const (
	DividerMilli int32 = 1000
	DividerUnit  int32 = 1
)

// MaxCells is the largest cell count the decoder will emit (matches the JK
// hardware's 24-cell status layout).
const MaxCells = 24

const cellTagBase uint16 = 0x7900

// CellTag returns the wire tag for cell n (1-based).
func CellTag(n int) uint16 { return cellTagBase + uint16(n) }

// Scalar field tags. The value equals the JK status data-ID so the tag stays
// stable and self-documenting.
const (
	TagTempMosfet    uint16 = 0x0080
	TagTempBox       uint16 = 0x0081
	TagTempBattery   uint16 = 0x0082
	TagTotalVoltage  uint16 = 0x0083
	TagCurrent       uint16 = 0x0084
	TagSOC           uint16 = 0x0085
	TagCycles        uint16 = 0x0087
	TagCycleCapacity uint16 = 0x0089
	TagCellCount     uint16 = 0x008A
	TagErrors        uint16 = 0x008B
	TagModes         uint16 = 0x008C
)
