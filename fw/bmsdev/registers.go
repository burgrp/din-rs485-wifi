// Package bmsdev holds the compiled register catalog for the JK BMS, shared by
// the bms firmware (for the published tag list) and the host hub (for building
// the provisioned device type).
package bmsdev

import "github.com/burgrp/din-rs485-wifi/fw/nwbms"

// ExposedCells is how many per-cell voltage registers are published. The JK
// B2A20S20P supports up to 20 series cells; packs with fewer cells leave the
// surplus registers null.
const ExposedCells = 20

// Scalar describes one non-cell BMS status field. Cell voltages are generated
// separately from ExposedCells.
type Scalar struct {
	Tag     uint16
	Name    string
	Title   string
	Unit    string
	Divider int32
	IsInt   bool
}

// Scalars are the non-cell status fields published to the hub. Voltages and
// currents use the milli divider; counts, SOC, temperatures and bitmasks are
// plain integers.
var Scalars = []Scalar{
	{nwbms.TagTotalVoltage, "voltage.total", "Total voltage", "V", nwbms.DividerMilli, false},
	{nwbms.TagCurrent, "current", "Current", "A", nwbms.DividerMilli, false},
	{nwbms.TagSOC, "soc", "State of charge", "%", nwbms.DividerUnit, false},
	{nwbms.TagTempMosfet, "temp.mosfet", "MOSFET temperature", "\u00B0C", nwbms.DividerUnit, false},
	{nwbms.TagTempBox, "temp.box", "Box temperature", "\u00B0C", nwbms.DividerUnit, false},
	{nwbms.TagTempBattery, "temp.battery", "Battery temperature", "\u00B0C", nwbms.DividerUnit, false},
	{nwbms.TagCycles, "cycles", "Charge cycles", "", nwbms.DividerUnit, true},
	{nwbms.TagCycleCapacity, "cycle.capacity", "Total cycle capacity", "Ah", nwbms.DividerUnit, false},
	{nwbms.TagCellCount, "cell.count", "Configured cell count", "", nwbms.DividerUnit, true},
	{nwbms.TagErrors, "errors", "Errors bitmask", "", nwbms.DividerUnit, true},
	{nwbms.TagModes, "modes", "Mode bitmask", "", nwbms.DividerUnit, true},
}

// Tags returns every wire tag the BMS firmware exposes, in a stable order: the
// scalar fields followed by the per-cell voltages.
func Tags() []uint16 {
	tags := make([]uint16, 0, len(Scalars)+ExposedCells)
	for i := range Scalars {
		tags = append(tags, Scalars[i].Tag)
	}
	for i := 1; i <= ExposedCells; i++ {
		tags = append(tags, nwbms.CellTag(i))
	}
	return tags
}
