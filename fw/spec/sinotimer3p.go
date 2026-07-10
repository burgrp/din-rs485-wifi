//go:build !tinygo

package spec

// SinotimerEnergyMeter3P is the built-in Sinotimer 3-phase energy meter type.
var SinotimerEnergyMeter3P = &Device{
	Name: "sinotimer-3p",
	Registers: []MeterRegister{
		{Name: "voltage.1", Title: "Voltage L1", Unit: "V", Address: 0x0000},
		{Name: "voltage.2", Title: "Voltage L2", Unit: "V", Address: 0x0002},
		{Name: "voltage.3", Title: "Voltage L3", Unit: "V", Address: 0x0004},
		{Name: "current.1", Title: "Current L1", Unit: "A", Address: 0x0008},
		{Name: "current.2", Title: "Current L2", Unit: "A", Address: 0x000A},
		{Name: "current.3", Title: "Current L3", Unit: "A", Address: 0x000C},
		{Name: "power.active.total", Title: "Total active power", Unit: "W", Address: 0x0010},
		{Name: "power.active.1", Title: "Active power L1", Unit: "W", Address: 0x0012},
		{Name: "power.active.2", Title: "Active power L2", Unit: "W", Address: 0x0014},
		{Name: "power.active.3", Title: "Active power L3", Unit: "W", Address: 0x0016},
		{Name: "power.reactive.total", Title: "Total reactive power", Unit: "VA", Address: 0x0018},
		{Name: "power.reactive.1", Title: "Reactive power L1", Unit: "VA", Address: 0x001A},
		{Name: "power.reactive.2", Title: "Reactive power L2", Unit: "VA", Address: 0x001C},
		{Name: "power.reactive.3", Title: "Reactive power L3", Unit: "VA", Address: 0x001E},
		{Name: "power.factor.1", Title: "Power factor L1", Unit: "", Address: 0x002A},
		{Name: "power.factor.2", Title: "Power factor L2", Unit: "", Address: 0x002C},
		{Name: "power.factor.3", Title: "Power factor L3", Unit: "", Address: 0x002E},
		{Name: "frequency", Title: "Frequency", Unit: "Hz", Address: 0x0036},
		{Name: "energy.active", Title: "Total active energy", Unit: "kWh", Address: 0x0100},
		{Name: "energy.reactive", Title: "Total reactive energy", Unit: "kVAh", Address: 0x0400},
	}}
