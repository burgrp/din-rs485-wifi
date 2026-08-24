//go:build !tinygo

package main

import (
	"github.com/burgrp/bleriot-rs485/fw-em/spec"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/site/cli"
)

var far = inventory.Channel{Name: "far", Number: 37, SpreadFactor: config.SpreadFactorS8}

func main() {
	meterType := spec.Type()
	meterConfig := spec.Config{
		MeterAddress: 1,
		Baud:         9600,
		Parity:       spec.ParityEven,
		WordOrder:    spec.WordOrderHighFirst,
		PollMs:       500,
		DisconnectMs: 5000,
	}
	cli.Start(inventory.Inventory{
		{
			Name:    "em.grid",
			Address: [4]byte{0x08, 0x1A, 0xBD, 0xC8},
			Key:     [16]byte{0x04, 0xD0, 0x98, 0x67, 0xFA, 0xE2, 0x22, 0x53, 0xE1, 0xFF, 0x5E, 0x69, 0xC9, 0xF5, 0x52, 0x6C},
			Channel: far,
			Type:    meterType,
			Config:  meterConfig,
		},
		{
			Name:    "em.house",
			Address: [4]byte{0x04, 0xF5, 0x06, 0x6E},
			Key:     [16]byte{0x20, 0x64, 0x53, 0x66, 0xC0, 0xEE, 0x8E, 0x53, 0xB8, 0xB0, 0xCF, 0x04, 0x78, 0x16, 0x1E, 0x64},
			Channel: far,
			Type:    meterType,
			Config:  meterConfig,
		},
	})
}
