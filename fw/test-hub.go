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
			UID:     [12]byte{0x5A, 0x33, 0x50, 0x41, 0x12, 0x32, 0x35, 0x32, 0x29, 0x93, 0x95, 0x00},
			Key:     [16]byte{0x63, 0x03, 0xF1, 0x7F, 0x54, 0x7C, 0xE6, 0x94, 0x70, 0xB2, 0x75, 0xD4, 0xB1, 0xA0, 0x65, 0x8E},
			Channel: channel,
			Type:    spec.TypeForSlaves(slaves),
			Config:  spec.ConfigForSlaves(slaves, 100),
		},
	})
}
