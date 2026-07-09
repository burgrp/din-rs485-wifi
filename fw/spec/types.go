package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

// Config is persisted in the provisioning page and read by firmware on boot.
type Config struct {
	GridAddr  uint8
	HouseAddr uint8
	PollMs    uint16
	Baud      uint32
}

func DefaultConfig() Config {
	return Config{
		GridAddr:  1,
		HouseAddr: 2,
		PollMs:    1000,
		Baud:      9600,
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
	Name      string
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

func Type() inventory.DeviceType {
	regs := make([]inventory.Register, 0, len(AllRegisters))
	for i := range AllRegisters {
		reg := AllRegisters[i]
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
