//go:build !tinygo

package bmsdev

import (
	"strconv"

	"github.com/burgrp/bleriot/lib/shared/inventory"

	"github.com/burgrp/din-rs485-wifi/fw/nwbms"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// Type builds the hub-visible device type from the compiled BMS catalog.
func Type() inventory.DeviceType {
	regs := make([]spec.Register, 0, len(Scalars)+ExposedCells)
	for i := range Scalars {
		s := Scalars[i]
		typ := inventory.TypeFloat
		if s.IsInt {
			typ = inventory.TypeInt
		}
		regs = append(regs, spec.Register{
			Tag:     s.Tag,
			Name:    s.Name,
			Title:   s.Title,
			Unit:    s.Unit,
			Type:    typ,
			Divider: s.Divider,
		})
	}
	for i := 1; i <= ExposedCells; i++ {
		n := strconv.Itoa(i)
		regs = append(regs, spec.Register{
			Tag:     nwbms.CellTag(i),
			Name:    "cell." + n,
			Title:   "Cell " + n + " voltage",
			Unit:    "V",
			Type:    inventory.TypeFloat,
			Divider: nwbms.DividerMilli,
		})
	}
	return spec.DeviceType("jk-bms", regs)
}
