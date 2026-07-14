//go:build !tinygo

// Command fw (host build) is the site/hub binary for the RS485 nodes. It
// provisions and serves the inventory: two Sinotimer 3-phase energy meters
// (grid and house) and one JK BMS. Each node runs its own compiled-in firmware;
// the hub supplies only identity, channel and poll interval.
package main

import (
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/site/cli"

	"github.com/burgrp/din-rs485-wifi/fw/meterdev"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

var channel = inventory.Channel{Name: "test", Number: 37, SpreadFactor: config.SpreadFactorS2}

func main() {
	meterType := meterdev.Type()
	cli.Start(inventory.Inventory{
		{
			Name:    "grid",
			UID:     [12]byte{0x5A, 0x33, 0x50, 0x41, 0x12, 0x32, 0x35, 0x32, 0x29, 0x93, 0x95, 0x00},
			Key:     [16]byte{0xD7, 0x0E, 0x5B, 0x73, 0x11, 0x22, 0x2D, 0xFE, 0xE8, 0x20, 0x84, 0x6D, 0x4D, 0x84, 0xC8, 0xD6},
			Channel: channel,
			Type:    meterType,
			Config:  spec.Config{PollMs: 1000},
		},
		// {
		// 	Name:    "house",
		// 	UID:     [12]byte{0x5A, 0x33, 0x50, 0x41, 0x12, 0x32, 0x35, 0x32, 0x29, 0x93, 0x95, 0x01},
		// 	Key:     [16]byte{0x2F, 0x9C, 0x84, 0x1B, 0xA7, 0x40, 0xD3, 0x66, 0x18, 0xE5, 0x0A, 0x92, 0x7B, 0xC4, 0x3D, 0x51},
		// 	Channel: channel,
		// 	Type:    meterType,
		// 	Config:  spec.Config{PollMs: 1000},
		// },
		// {
		// 	Name:    "bms",
		// 	UID:     [12]byte{0x5A, 0x33, 0x42, 0x4D, 0x53, 0x01, 0x35, 0x32, 0x29, 0x93, 0x95, 0x00},
		// 	Key:     [16]byte{0x7A, 0x11, 0xC4, 0x2E, 0x58, 0x9B, 0x30, 0x6D, 0xA1, 0x4F, 0x88, 0xE2, 0x1C, 0x77, 0x05, 0x93},
		// 	Channel: channel,
		// 	Type:    bmsdev.Type(),
		// 	Config:  spec.Config{PollMs: 1000},
		// },
	})
}
