package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

// Config is persisted in the provisioning page and read by firmware on boot.
type Config struct {
	SlaveAddr1 uint8
	SlaveAddr2 uint8
	PollMs     uint16
	Baud       uint32
}

func DefaultConfig() Config {
	return Config{
		SlaveAddr1: 1,
		SlaveAddr2: 2,
		PollMs:     1000,
		Baud:       9600,
	}
}

// RegDef defines one wire-visible register mapped from a Modbus address.
type RegDef struct {
	Tag        uint8
	Name       string
	Title      string
	Unit       string
	Address    uint16
	ScaleMilli int32
}

// DeviceDef describes one RS485 slave to poll.
type DeviceDef struct {
	SlaveAddr uint8
	Baud      uint32
	Registers []RegDef
}

// NodeConfig contains node-local pin assignments and runtime settings.
type NodeConfig struct {
	RS485TxPin   string
	RS485RxPin   string
	RS485TxEnPin string
	LedPin       string
	DebugPin     string
	RFDataPin    string
	RFSckPin     string
	RFCsPin      string
	ReadRetries  int
}

var Chip = inventory.PY32F003x6

// DeviceBinding maps one RS485 slave type into the bridge inventory namespace.
// TagOffset is explicit so different device register maps can coexist safely.
type DeviceBinding struct {
	Group     string
	Slot      uint8
	TagOffset uint8
	Registers []RegDef
}

func TypeForGroups(groups []string) inventory.DeviceType {
	groupSlots := make([]GroupSlot, 0, len(groups))
	for i := range groups {
		groupSlots = append(groupSlots, GroupSlot{Group: groups[i], Slot: uint8(i)})
	}
	return TypeForGroupSlots(groupSlots)
}

func TypeForGroupSlots(groupSlots []GroupSlot) inventory.DeviceType {
	regs := make([]inventory.Register, 0, len(RegistersForGroupSlots(groupSlots)))
	for _, reg := range RegistersForGroupSlots(groupSlots) {
		divider := reg.ScaleMilli
		if divider <= 0 {
			divider = 1
		}
		md := map[string]string{"title": reg.Title}
		if reg.Unit != "" {
			md["unit"] = reg.Unit
		}
		regs = append(regs, inventory.Register{
			Tag:        reg.Tag,
			Name:       reg.Name,
			Type:       inventory.TypeFloat,
			Multiplier: 1,
			Divider:    divider,
			Metadata:   md,
		})
	}

	return inventory.DeviceType{
		Name:      "din-rs485-bridge",
		Chip:      Chip,
		Registers: regs,
	}
}

func TypeForGroupBindings(groupBindings []GroupBinding) inventory.DeviceType {
	regs := make([]inventory.Register, 0, len(RegistersForGroupBindings(groupBindings)))
	for _, reg := range RegistersForGroupBindings(groupBindings) {
		divider := reg.ScaleMilli
		if divider <= 0 {
			divider = 1
		}
		md := map[string]string{"title": reg.Title}
		if reg.Unit != "" {
			md["unit"] = reg.Unit
		}
		regs = append(regs, inventory.Register{
			Tag:        reg.Tag,
			Name:       reg.Name,
			Type:       inventory.TypeFloat,
			Multiplier: 1,
			Divider:    divider,
			Metadata:   md,
		})
	}

	return inventory.DeviceType{
		Name:      "din-rs485-bridge",
		Chip:      Chip,
		Registers: regs,
	}
}

// RegistersForDeviceBindings builds host-visible register definitions for mixed RS485 device types.
func RegistersForDeviceBindings(bindings []DeviceBinding) []RegDef {
	total := 0
	for i := range bindings {
		total += len(bindings[i].Registers)
	}
	out := make([]RegDef, 0, total)
	for i := range bindings {
		prefix := bindings[i].Group
		if prefix != "" {
			prefix += "."
		}
		for j := range bindings[i].Registers {
			reg := bindings[i].Registers[j]
			reg.Tag = addTagOffset(reg.Tag, bindings[i].TagOffset)
			if prefix != "" {
				reg.Name = prefix + reg.Name
			}
			out = append(out, reg)
		}
	}
	return out
}

func TypeForDeviceBindings(bindings []DeviceBinding) inventory.DeviceType {
	regs := make([]inventory.Register, 0, len(RegistersForDeviceBindings(bindings)))
	for _, reg := range RegistersForDeviceBindings(bindings) {
		divider := reg.ScaleMilli
		if divider <= 0 {
			divider = 1
		}
		md := map[string]string{"title": reg.Title}
		if reg.Unit != "" {
			md["unit"] = reg.Unit
		}
		regs = append(regs, inventory.Register{
			Tag:        reg.Tag,
			Name:       reg.Name,
			Type:       inventory.TypeFloat,
			Multiplier: 1,
			Divider:    divider,
			Metadata:   md,
		})
	}

	return inventory.DeviceType{
		Name:      "din-rs485-bridge",
		Chip:      Chip,
		Registers: regs,
	}
}
