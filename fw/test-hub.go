//go:build !tinygo

package main

import (
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/site/cli"

	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

var channel = inventory.Channel{Name: "far", Number: 37, SpreadFactor: config.SpreadFactorS8}
var slaves = []spec.Slave{
	{Group: "grid", Addr: 1, TagRegion: 1, Device: spec.SinotimerEnergyMeter3P},
	{Group: "house", Addr: 2, TagRegion: 2, Device: spec.SinotimerEnergyMeter3P},
}

func main() {
	cli.Start(inventory.Inventory{
		{
			Name:    "em",
			UID:     [12]byte{0x44, 0x49, 0x4E, 0x2D, 0x52, 0x53, 0x34, 0x38, 0x35, 0x00, 0x00, 0x01},
			Key:     [16]byte{0x31, 0x5A, 0x4B, 0x0C, 0x77, 0x3E, 0x11, 0xE2, 0xA6, 0x99, 0xB4, 0x20, 0x4D, 0x12, 0x61, 0x8F},
			Channel: channel,
			Type:    spec.TypeForSlaves(slaves),
			Config:  spec.ConfigForSlaves(slaves, 1000),
		},
	})
}
