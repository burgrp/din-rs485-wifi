package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

// Chip is the provisioning target for the DIN RS485 bridge.
var Chip = inventory.PY32F003x6

const (
	// MaxRuns / MaxSlaves bound the fixed-size provisioning payload so it fits
	// in the tiny PY32 flash config page (~30 usable bytes). Runs form a shared
	// pool; each slave references a slice of it, so identical device types cost
	// pool space only once.
	MaxRuns   = 10
	MaxSlaves = 2

	// tagAddrBits splits the 16-bit wire tag into a slave selector and a Modbus
	// word address: tag = sel<<tagAddrBits | addr. 12 address bits cover the
	// full 0..4095 register range; 4 selector bits cover slaves 1..15.
	tagAddrBits = 12
	tagAddrMask = uint16(1)<<tagAddrBits - 1

	// runCountBits is how many low bits of a packed Run hold the register count.
	// The remaining high bits hold the base word address, which is why a single
	// run may span at most MaxRunAddr and MaxRunCount.
	runCountBits = 4
	runCountMask = uint16(1)<<runCountBits - 1
	// MaxRunAddr is the largest Modbus word address a run (or tag) can encode.
	MaxRunAddr = uint16(1)<<tagAddrBits - 1
	// MaxRunCount is the largest register count a single run can hold. Longer
	// contiguous spans must be split across multiple runs.
	MaxRunCount = uint8(1)<<runCountBits - 1

	// WireScaleMilli is the fixed integer scale applied to every float register
	// before it goes on the wire. The hub divides by the same factor. Uniform
	// scaling keeps the device fully generic (no per-register scale knowledge).
	WireScaleMilli = 1000
)

// Run packs a contiguous block of float32 registers into 16 bits:
// bits 15..4 = base Modbus word address (0..MaxRunAddr), bits 3..0 = register
// count (0..MaxRunCount). Register j of the run lives at Addr + j*2, because
// each float32 spans two Modbus words.
type Run uint16

// MakeRun packs a base address and register count into a Run.
func MakeRun(addr uint16, count uint8) Run {
	return Run(addr<<runCountBits | uint16(count)&runCountMask)
}

// Addr returns the run's base Modbus word address.
func (r Run) Addr() uint16 { return uint16(r) >> runCountBits }

// Count returns the number of float32 registers in the run.
func (r Run) Count() uint8 { return uint8(uint16(r) & runCountMask) }

// SlaveCfg binds an RS485 Modbus slave to a tag selector and a slice of the
// shared run pool. Addr == 0 marks the slot unused.
type SlaveCfg struct {
	// Addr is the RS485 Modbus slave address to poll (0 = slot unused).
	Addr uint8
	// Sel is the wire-tag selector (1..15) for this slave's registers.
	Sel uint8
	// Span packs the slave's run slice: high nibble = first run index, low
	// nibble = run count. Use RunStart / RunCount to read it.
	Span uint8
}

// MakeSpan packs a run-pool slice (start index, length) into a Span byte.
func MakeSpan(start, count uint8) uint8 { return start<<4 | count&0x0F }

// RunStart returns the index of this slave's first run in the shared pool.
func (s SlaveCfg) RunStart() uint8 { return s.Span >> 4 }

// RunCount returns how many pool runs belong to this slave.
func (s SlaveCfg) RunCount() uint8 { return s.Span & 0x0F }

// Config is the fixed-size provisioning payload persisted in the flash page.
// It carries no meter semantics: only which Modbus addresses to poll on which
// slaves, and how the results map onto wire tags via Sel.
//
// Layout (little-endian, no padding): PollMs(2) + Runs(10*2=20) +
// Slaves(2*3=6) = 28 bytes.
type Config struct {
	PollMs uint16
	Runs   [MaxRuns]Run
	Slaves [MaxSlaves]SlaveCfg
}

// TagFor encodes a slave selector and a Modbus word address into a wire tag.
// sel must be 1..15 (0 is reserved by the protocol as RegAll).
func TagFor(sel uint8, addr uint16) uint16 {
	return uint16(sel)<<tagAddrBits | (addr & tagAddrMask)
}
