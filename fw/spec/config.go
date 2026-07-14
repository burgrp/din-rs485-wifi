// Package spec holds the small, hardware-facing facts shared between a device
// firmware and its host-side provisioning tool: the provisioning chip profile,
// the fixed provisioning payload, and the uniform wire scale. Register and
// field tables are device-specific and live in each firmware's own package.
package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

// Chip is the provisioning target. The firmware runs on the PY32F003x6 (32k
// flash / 4k RAM); the PY32F030x8 profile (inventory.PY32F030x8) is available
// for development headroom if needed.
var Chip = inventory.PY32F003x6

// WireScaleMilli is the default fixed integer scale applied to float registers
// before they go on the wire; the hub divides by the same factor. A register
// may override it with its own divider when a different resolution is wanted.
const WireScaleMilli = 1000

// Config is the fixed-size provisioning payload persisted in the flash page.
// With one slave per node the firmware needs no addressing map from the hub:
// the register/field table is compiled in, so the only runtime knob is the
// poll interval.
//
// Layout (little-endian, no padding): PollMs(2) = 2 bytes.
type Config struct {
	PollMs uint16
}

// DecodeConfig parses the provisioning payload. It mirrors
// encoding/binary.Write over Config exactly and tolerates a short buffer.
func DecodeConfig(raw []byte) Config {
	var c Config
	if len(raw) >= 2 {
		c.PollMs = uint16(raw[0]) | uint16(raw[1])<<8
	}
	return c
}
