//go:build !tinygo

package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

// Register is the host-side description of one device register or field. It is
// the common shape both firmwares reduce their compiled-in tables to when
// building the hub-visible device type.
type Register struct {
	// Tag is the permanent wire identity (meter = Modbus word address, BMS =
	// protocol data ID). It must match the tag the firmware notifies.
	Tag uint16
	// Name is the register name exposed to the hub and Registry.
	Name string
	// Title and Unit are descriptive metadata forwarded to the Registry.
	Title string
	Unit  string
	// Type interprets the int32 wire value. Empty selects inventory.TypeFloat.
	Type inventory.RegType
	// Divider scales the wire value for display (display = wire / Divider). Zero
	// selects WireScaleMilli.
	Divider int32
}

// DeviceType builds the hub-visible device type from a compiled register table.
func DeviceType(name string, regs []Register) inventory.DeviceType {
	out := make([]inventory.Register, 0, len(regs))
	for i := range regs {
		r := regs[i]
		typ := r.Type
		if typ == "" {
			typ = inventory.TypeFloat
		}
		div := r.Divider
		if div == 0 {
			div = WireScaleMilli
		}
		md := map[string]string{}
		if r.Title != "" {
			md["title"] = r.Title
		}
		if r.Unit != "" {
			md["unit"] = r.Unit
		}
		out = append(out, inventory.Register{
			Tag:        r.Tag,
			Name:       r.Name,
			Type:       typ,
			Multiplier: 1,
			Divider:    div,
			Metadata:   md,
		})
	}
	return inventory.DeviceType{Name: name, Chip: Chip, Registers: out}
}
