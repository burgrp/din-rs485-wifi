// Package meterdev holds the compiled register table for the Sinotimer 3-phase
// energy meter, shared by the meter firmware (for polling) and the host hub (for
// building the provisioned device type).
package meterdev

// Register describes one Modbus register of the energy meter. The wire tag is
// the Modbus word address; every value is a float32 spanning two words, so
// consecutive registers are spaced by 2.
type Register struct {
	Name    string
	Title   string
	Unit    string
	Address uint16
}

// Registers is the built-in Sinotimer 3-phase energy meter table.
var Registers = []Register{
	{Name: "voltage.1", Title: "Voltage L1", Unit: "V", Address: 0x0000},
	{Name: "voltage.2", Title: "Voltage L2", Unit: "V", Address: 0x0002},
	{Name: "voltage.3", Title: "Voltage L3", Unit: "V", Address: 0x0004},
	{Name: "current.1", Title: "Current L1", Unit: "A", Address: 0x0008},
	{Name: "current.2", Title: "Current L2", Unit: "A", Address: 0x000A},
	{Name: "current.3", Title: "Current L3", Unit: "A", Address: 0x000C},
	{Name: "power.active.total", Title: "Total active power", Unit: "W", Address: 0x0010},
	{Name: "power.active.1", Title: "Active power L1", Unit: "W", Address: 0x0012},
	{Name: "power.active.2", Title: "Active power L2", Unit: "W", Address: 0x0014},
	{Name: "power.active.3", Title: "Active power L3", Unit: "W", Address: 0x0016},
	{Name: "power.reactive.total", Title: "Total reactive power", Unit: "VA", Address: 0x0018},
	{Name: "power.reactive.1", Title: "Reactive power L1", Unit: "VA", Address: 0x001A},
	{Name: "power.reactive.2", Title: "Reactive power L2", Unit: "VA", Address: 0x001C},
	{Name: "power.reactive.3", Title: "Reactive power L3", Unit: "VA", Address: 0x001E},
	{Name: "power.factor.1", Title: "Power factor L1", Unit: "", Address: 0x002A},
	{Name: "power.factor.2", Title: "Power factor L2", Unit: "", Address: 0x002C},
	{Name: "power.factor.3", Title: "Power factor L3", Unit: "", Address: 0x002E},
	{Name: "frequency", Title: "Frequency", Unit: "Hz", Address: 0x0036},
	{Name: "energy.active", Title: "Total active energy", Unit: "kWh", Address: 0x0100},
	{Name: "energy.reactive", Title: "Total reactive energy", Unit: "kVAh", Address: 0x0400},
}

// TagBase offsets a Modbus word address into an RF register tag. Tag 0 is
// reserved by the inventory protocol (WatchAll subscribes to register ID 0), so
// voltage.1 at Modbus address 0x0000 would collide; the +1 offset keeps every
// tag non-zero while preserving the addresses' one-to-one order.
const TagBase = 1

// Tag maps a Modbus word address to the register's permanent RF tag. Both the
// firmware (which notifies) and the host device type must use it so the tags
// agree on the wire.
func Tag(addr uint16) uint16 { return addr + TagBase }

// MaxRunRegs caps a single Modbus read at 15 float32 registers (30 words), the
// rs485 client's receive-buffer limit. Longer contiguous spans are split.
const MaxRunRegs = 15

// Run is one Modbus read: Count float32 registers starting at Base.
type Run struct {
	Base  uint16
	Count uint8
}

// Runs compresses the register table into contiguous reads (stride 2 words),
// splitting spans longer than MaxRunRegs.
func Runs() []Run {
	runs := make([]Run, 0, len(Registers))
	for i := range Registers {
		addr := Registers[i].Address
		if n := len(runs); n > 0 {
			last := &runs[n-1]
			if addr == last.Base+uint16(last.Count)*2 && last.Count < MaxRunRegs {
				last.Count++
				continue
			}
		}
		runs = append(runs, Run{Base: addr, Count: 1})
	}
	return runs
}
