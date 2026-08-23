//go:build !tinygo

package spec

import (
	"fmt"
	"math"

	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/shared/puya"
)

var Chip = puya.PY32F030x8

func Type() inventory.DeviceType {
	register := func(tag uint16, name, title, unit string) inventory.Register {
		return inventory.Register{
			Tag:      tag,
			Name:     name,
			Type:     inventory.TypeFloat,
			ReadOnly: true,
			Conversion: inventory.Conversion{
				Decode: decodeFloat32,
			},
			Metadata: map[string]string{"title": title, "unit": unit},
		}
	}

	return inventory.DeviceType{
		Name: "em",
		Chip: Chip,
		Registers: []inventory.Register{
			register(RegVoltage1, "voltage.1", "Voltage L1", "V"),
			register(RegVoltage2, "voltage.2", "Voltage L2", "V"),
			register(RegVoltage3, "voltage.3", "Voltage L3", "V"),
			register(RegCurrent1, "current.1", "Current L1", "A"),
			register(RegCurrent2, "current.2", "Current L2", "A"),
			register(RegCurrent3, "current.3", "Current L3", "A"),
			register(RegPowerActiveTotal, "power.active.total", "Total active power", "W"),
			register(RegPowerActive1, "power.active.1", "Active power L1", "W"),
			register(RegPowerActive2, "power.active.2", "Active power L2", "W"),
			register(RegPowerActive3, "power.active.3", "Active power L3", "W"),
			register(RegPowerReactiveTotal, "power.reactive.total", "Total reactive power", "VA"),
			register(RegPowerReactive1, "power.reactive.1", "Reactive power L1", "VA"),
			register(RegPowerReactive2, "power.reactive.2", "Reactive power L2", "VA"),
			register(RegPowerReactive3, "power.reactive.3", "Reactive power L3", "VA"),
			register(RegPowerFactor1, "power.factor.1", "Power factor L1", ""),
			register(RegPowerFactor2, "power.factor.2", "Power factor L2", ""),
			register(RegPowerFactor3, "power.factor.3", "Power factor L3", ""),
			register(RegFrequency, "frequency", "Frequency", "Hz"),
			register(RegEnergyActive, "energy.active", "Total active energy", "kWh"),
			register(RegEnergyReactive, "energy.reactive", "Total reactive energy", "kVAh"),
		},
	}
}

func decodeFloat32(raw int32) (any, error) {
	value := math.Float32frombits(uint32(raw))
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return nil, fmt.Errorf("invalid meter float %08x", uint32(raw))
	}
	return float64(value), nil
}
