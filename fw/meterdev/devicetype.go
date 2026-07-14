//go:build !tinygo

package meterdev

import (
	"github.com/burgrp/bleriot/lib/shared/inventory"

	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// Type builds the hub-visible device type from the compiled meter table.
func Type() inventory.DeviceType {
	regs := make([]spec.Register, len(Registers))
	for i := range Registers {
		r := Registers[i]
		regs[i] = spec.Register{
			Tag:   Tag(r.Address),
			Name:  r.Name,
			Title: r.Title,
			Unit:  r.Unit,
		}
	}
	return spec.DeviceType("sinotimer-3p", regs)
}
