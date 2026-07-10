//go:build !tinygo

package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

// MeterRegister is the hub-side description of one Modbus register. This table
// lives only on the hub: the device firmware never sees names, titles, units,
// or scales. Addresses are Modbus word addresses; every value is float32 (two
// words) so consecutive entries are spaced by 2.
type MeterRegister struct {
	Name    string
	Title   string
	Unit    string
	Address uint16
}

// Device is a hub-side Modbus device type: a named register table. Different
// slaves may reference different Devices, so an energy meter and some other
// kind of Modbus device can share one bridge.
type Device struct {
	Name      string
	Registers []MeterRegister
}

// Slave maps a hub-side group name to a physical RS485 slave (Modbus address),
// a wire-tag region, and the Modbus device type wired to that slave. Mixing
// device types is supported: each slave carries its own register table.
type Slave struct {
	Group     string
	Addr      uint8
	TagRegion uint8
	Device    *Device
}

// runs compresses a device's register table into packed contiguous runs
// (stride 2 words). Runs longer than MaxRunCount are split; addresses above
// MaxRunAddr are rejected because they cannot be encoded in a Run or wire tag.
func (d *Device) runs() []Run {
	runs := make([]Run, 0, MaxRuns)
	for i := range d.Registers {
		addr := d.Registers[i].Address
		if addr > MaxRunAddr {
			panic("spec: register address exceeds encodable range")
		}
		if n := len(runs); n > 0 {
			last := runs[n-1]
			if addr == last.Addr()+uint16(last.Count())*2 && last.Count() < MaxRunCount {
				runs[n-1] = MakeRun(last.Addr(), last.Count()+1)
				continue
			}
		}
		runs = append(runs, MakeRun(addr, 1))
	}
	return runs
}

// ConfigForSlaves builds the provisioning payload for the given slaves. Runs go
// into a shared pool; slaves of the same Device share one pool slice, so two
// identical meters cost pool space only once. Panics if the combined distinct
// runs exceed MaxRuns.
func ConfigForSlaves(slaves []Slave, pollMs uint16) Config {
	cfg := Config{PollMs: pollMs}
	pool := make([]Run, 0, MaxRuns)
	ranges := make(map[*Device][2]uint8)
	for i := range slaves {
		if i >= MaxSlaves {
			panic("spec: more slaves than MaxSlaves")
		}
		sl := slaves[i]
		rng, ok := ranges[sl.Device]
		if !ok {
			runs := sl.Device.runs()
			start := len(pool)
			pool = append(pool, runs...)
			if len(pool) > MaxRuns {
				panic("spec: combined runs exceed MaxRuns")
			}
			rng = [2]uint8{uint8(start), uint8(len(runs))}
			ranges[sl.Device] = rng
		}
		cfg.Slaves[i] = SlaveCfg{Addr: sl.Addr, TagRegion: sl.TagRegion, Span: MakeSpan(rng[0], rng[1])}
	}
	for i := range pool {
		cfg.Runs[i] = pool[i]
	}
	return cfg
}

// TypeForSlaves builds the hub-visible device type. Each register's wire tag is
// TagFor(slave.TagRegion, address); names are prefixed with the group, e.g.
// "grid.voltage.1". Every register uses the uniform WireScaleMilli divider.
func TypeForSlaves(slaves []Slave) inventory.DeviceType {
	total := 0
	for i := range slaves {
		total += len(slaves[i].Device.Registers)
	}
	regs := make([]inventory.Register, 0, total)
	for i := range slaves {
		sl := slaves[i]
		prefix := sl.Group
		if prefix != "" {
			prefix += "."
		}
		for j := range sl.Device.Registers {
			m := sl.Device.Registers[j]
			md := map[string]string{"title": m.Title}
			if m.Unit != "" {
				md["unit"] = m.Unit
			}
			regs = append(regs, inventory.Register{
				Tag:        TagFor(sl.TagRegion, m.Address),
				Name:       prefix + m.Name,
				Type:       inventory.TypeFloat,
				Multiplier: 1,
				Divider:    WireScaleMilli,
				Metadata:   md,
			})
		}
	}
	return inventory.DeviceType{
		Name:      "din-rs485-bridge",
		Chip:      Chip,
		Registers: regs,
	}
}
