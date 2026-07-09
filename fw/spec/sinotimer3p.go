package spec

// Keep tags stable once deployed. Do not renumber shipped tags.
const (
	TagVoltageL1 uint8 = iota + 1
	TagVoltageL2
	TagVoltageL3
	TagCurrentL1
	TagCurrentL2
	TagCurrentL3
	TagPowerActiveTotal
	TagPowerActiveL1
	TagPowerActiveL2
	TagPowerActiveL3
	TagPowerReactiveTotal
	TagPowerReactiveL1
	TagPowerReactiveL2
	TagPowerReactiveL3
	TagPowerFactorL1
	TagPowerFactorL2
	TagPowerFactorL3
	TagFrequency
	TagEnergyActive
	TagEnergyReactive
)

// EnergyMeterRegisters is the canonical Sinotimer 3P definition.
// Names intentionally match the original MicroPython register names.
var EnergyMeterRegisters = []RegDef{
	{Tag: TagVoltageL1, Name: "voltage.1", Title: "Voltage L1", Unit: "V", Address: 0x0000, ScaleMilli: 1000},
	{Tag: TagVoltageL2, Name: "voltage.2", Title: "Voltage L2", Unit: "V", Address: 0x0002, ScaleMilli: 1000},
	{Tag: TagVoltageL3, Name: "voltage.3", Title: "Voltage L3", Unit: "V", Address: 0x0004, ScaleMilli: 1000},
	{Tag: TagCurrentL1, Name: "current.1", Title: "Current L1", Unit: "A", Address: 0x0008, ScaleMilli: 1000},
	{Tag: TagCurrentL2, Name: "current.2", Title: "Current L2", Unit: "A", Address: 0x000A, ScaleMilli: 1000},
	{Tag: TagCurrentL3, Name: "current.3", Title: "Current L3", Unit: "A", Address: 0x000C, ScaleMilli: 1000},
	{Tag: TagPowerActiveTotal, Name: "power.active.total", Title: "Total active power", Unit: "W", Address: 0x0010, ScaleMilli: 1},
	{Tag: TagPowerActiveL1, Name: "power.active.1", Title: "Active power L1", Unit: "W", Address: 0x0012, ScaleMilli: 1},
	{Tag: TagPowerActiveL2, Name: "power.active.2", Title: "Active power L2", Unit: "W", Address: 0x0014, ScaleMilli: 1},
	{Tag: TagPowerActiveL3, Name: "power.active.3", Title: "Active power L3", Unit: "W", Address: 0x0016, ScaleMilli: 1},
	{Tag: TagPowerReactiveTotal, Name: "power.reactive.total", Title: "Total reactive power", Unit: "VA", Address: 0x0018, ScaleMilli: 1},
	{Tag: TagPowerReactiveL1, Name: "power.reactive.1", Title: "Reactive power L1", Unit: "VA", Address: 0x001A, ScaleMilli: 1},
	{Tag: TagPowerReactiveL2, Name: "power.reactive.2", Title: "Reactive power L2", Unit: "VA", Address: 0x001C, ScaleMilli: 1},
	{Tag: TagPowerReactiveL3, Name: "power.reactive.3", Title: "Reactive power L3", Unit: "VA", Address: 0x001E, ScaleMilli: 1},
	{Tag: TagPowerFactorL1, Name: "power.factor.1", Title: "Power factor L1", Unit: "", Address: 0x002A, ScaleMilli: 1000},
	{Tag: TagPowerFactorL2, Name: "power.factor.2", Title: "Power factor L2", Unit: "", Address: 0x002C, ScaleMilli: 1000},
	{Tag: TagPowerFactorL3, Name: "power.factor.3", Title: "Power factor L3", Unit: "", Address: 0x002E, ScaleMilli: 1000},
	{Tag: TagFrequency, Name: "frequency", Title: "Frequency", Unit: "Hz", Address: 0x0036, ScaleMilli: 1000},
	{Tag: TagEnergyActive, Name: "energy.active", Title: "Total active energy", Unit: "kWh", Address: 0x0100, ScaleMilli: 1000},
	{Tag: TagEnergyReactive, Name: "energy.reactive", Title: "Total reactive energy", Unit: "kVAh", Address: 0x0400, ScaleMilli: 1000},
}

func RegistersPerMeter() uint8 {
	return uint8(len(EnergyMeterRegisters))
}

// RegistersForMeterSlot returns the register map for one RS485 meter slot.
// slot=0 uses base tags; slot=1 is offset by RegistersPerMeter(), etc.
func RegistersForMeterSlot(slot uint8) []RegDef {
	return registersForSlot(slot, "")
}

func RegistersForSlots(count uint8) []RegDef {
	out := make([]RegDef, 0, len(EnergyMeterRegisters)*int(count))
	for slot := uint8(0); slot < count; slot++ {
		out = append(out, RegistersForMeterSlot(slot)...)
	}
	return out
}

// GroupSlot explicitly maps a hub-side group name to an RS485 slot index.
type GroupSlot struct {
	Group string
	Slot  uint8
}

// RegistersForGroups builds host-visible register definitions for multiple hub-side groups.
// Group names are prefixed to base names, e.g. "grid" + "voltage.1" -> "grid.voltage.1".
func RegistersForGroups(groups []string) []RegDef {
	out := make([]RegDef, 0, len(EnergyMeterRegisters)*len(groups))
	for i := range groups {
		prefix := groups[i]
		if prefix != "" {
			prefix += "."
		}
		out = append(out, registersForSlot(uint8(i), prefix)...)
	}
	return out
}

// RegistersForGroupSlots builds host-visible register definitions from explicit group->slot mapping.
func RegistersForGroupSlots(groups []GroupSlot) []RegDef {
	out := make([]RegDef, 0, len(EnergyMeterRegisters)*len(groups))
	for i := range groups {
		prefix := groups[i].Group
		if prefix != "" {
			prefix += "."
		}
		out = append(out, registersForSlot(groups[i].Slot, prefix)...)
	}
	return out
}

func registersForSlot(slot uint8, namePrefix string) []RegDef {
	out := make([]RegDef, len(EnergyMeterRegisters))
	offset := uint16(slot) * uint16(len(EnergyMeterRegisters))
	for i := range EnergyMeterRegisters {
		reg := EnergyMeterRegisters[i]
		reg.Tag = uint8(uint16(reg.Tag) + offset)
		if namePrefix != "" {
			reg.Name = namePrefix + reg.Name
		}
		out[i] = reg
	}
	return out
}
