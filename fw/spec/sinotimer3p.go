package spec

// Keep tags stable once deployed. Do not renumber shipped tags.
const (
	TagGridVoltageL1 uint8 = iota + 1
	TagGridVoltageL2
	TagGridVoltageL3
	TagGridCurrentL1
	TagGridCurrentL2
	TagGridCurrentL3
	TagGridPowerActiveTotal
	TagGridPowerActiveL1
	TagGridPowerActiveL2
	TagGridPowerActiveL3
	TagGridPowerReactiveTotal
	TagGridPowerReactiveL1
	TagGridPowerReactiveL2
	TagGridPowerReactiveL3
	TagGridPowerFactorL1
	TagGridPowerFactorL2
	TagGridPowerFactorL3
	TagGridFrequency
	TagGridEnergyActive
	TagGridEnergyReactive

	TagHouseVoltageL1
	TagHouseVoltageL2
	TagHouseVoltageL3
	TagHouseCurrentL1
	TagHouseCurrentL2
	TagHouseCurrentL3
	TagHousePowerActiveTotal
	TagHousePowerActiveL1
	TagHousePowerActiveL2
	TagHousePowerActiveL3
	TagHousePowerReactiveTotal
	TagHousePowerReactiveL1
	TagHousePowerReactiveL2
	TagHousePowerReactiveL3
	TagHousePowerFactorL1
	TagHousePowerFactorL2
	TagHousePowerFactorL3
	TagHouseFrequency
	TagHouseEnergyActive
	TagHouseEnergyReactive
)

var GridRegisters = []RegDef{
	{Tag: TagGridVoltageL1, Name: "grid_voltage_1", Title: "Grid voltage L1", Unit: "V", Address: 0x0000, ScaleMilli: 1000},
	{Tag: TagGridVoltageL2, Name: "grid_voltage_2", Title: "Grid voltage L2", Unit: "V", Address: 0x0002, ScaleMilli: 1000},
	{Tag: TagGridVoltageL3, Name: "grid_voltage_3", Title: "Grid voltage L3", Unit: "V", Address: 0x0004, ScaleMilli: 1000},
	{Tag: TagGridCurrentL1, Name: "grid_current_1", Title: "Grid current L1", Unit: "A", Address: 0x0008, ScaleMilli: 1000},
	{Tag: TagGridCurrentL2, Name: "grid_current_2", Title: "Grid current L2", Unit: "A", Address: 0x000A, ScaleMilli: 1000},
	{Tag: TagGridCurrentL3, Name: "grid_current_3", Title: "Grid current L3", Unit: "A", Address: 0x000C, ScaleMilli: 1000},
	{Tag: TagGridPowerActiveTotal, Name: "grid_power_active_total", Title: "Grid total active power", Unit: "W", Address: 0x0010, ScaleMilli: 1},
	{Tag: TagGridPowerActiveL1, Name: "grid_power_active_1", Title: "Grid active power L1", Unit: "W", Address: 0x0012, ScaleMilli: 1},
	{Tag: TagGridPowerActiveL2, Name: "grid_power_active_2", Title: "Grid active power L2", Unit: "W", Address: 0x0014, ScaleMilli: 1},
	{Tag: TagGridPowerActiveL3, Name: "grid_power_active_3", Title: "Grid active power L3", Unit: "W", Address: 0x0016, ScaleMilli: 1},
	{Tag: TagGridPowerReactiveTotal, Name: "grid_power_reactive_total", Title: "Grid total reactive power", Unit: "VA", Address: 0x0018, ScaleMilli: 1},
	{Tag: TagGridPowerReactiveL1, Name: "grid_power_reactive_1", Title: "Grid reactive power L1", Unit: "VA", Address: 0x001A, ScaleMilli: 1},
	{Tag: TagGridPowerReactiveL2, Name: "grid_power_reactive_2", Title: "Grid reactive power L2", Unit: "VA", Address: 0x001C, ScaleMilli: 1},
	{Tag: TagGridPowerReactiveL3, Name: "grid_power_reactive_3", Title: "Grid reactive power L3", Unit: "VA", Address: 0x001E, ScaleMilli: 1},
	{Tag: TagGridPowerFactorL1, Name: "grid_power_factor_1", Title: "Grid power factor L1", Unit: "", Address: 0x002A, ScaleMilli: 1000},
	{Tag: TagGridPowerFactorL2, Name: "grid_power_factor_2", Title: "Grid power factor L2", Unit: "", Address: 0x002C, ScaleMilli: 1000},
	{Tag: TagGridPowerFactorL3, Name: "grid_power_factor_3", Title: "Grid power factor L3", Unit: "", Address: 0x002E, ScaleMilli: 1000},
	{Tag: TagGridFrequency, Name: "grid_frequency", Title: "Grid frequency", Unit: "Hz", Address: 0x0036, ScaleMilli: 1000},
	{Tag: TagGridEnergyActive, Name: "grid_energy_active", Title: "Grid total active energy", Unit: "kWh", Address: 0x0100, ScaleMilli: 1000},
	{Tag: TagGridEnergyReactive, Name: "grid_energy_reactive", Title: "Grid total reactive energy", Unit: "kVAh", Address: 0x0400, ScaleMilli: 1000},
}

var HouseRegisters = []RegDef{
	{Tag: TagHouseVoltageL1, Name: "house_voltage_1", Title: "House voltage L1", Unit: "V", Address: 0x0000, ScaleMilli: 1000},
	{Tag: TagHouseVoltageL2, Name: "house_voltage_2", Title: "House voltage L2", Unit: "V", Address: 0x0002, ScaleMilli: 1000},
	{Tag: TagHouseVoltageL3, Name: "house_voltage_3", Title: "House voltage L3", Unit: "V", Address: 0x0004, ScaleMilli: 1000},
	{Tag: TagHouseCurrentL1, Name: "house_current_1", Title: "House current L1", Unit: "A", Address: 0x0008, ScaleMilli: 1000},
	{Tag: TagHouseCurrentL2, Name: "house_current_2", Title: "House current L2", Unit: "A", Address: 0x000A, ScaleMilli: 1000},
	{Tag: TagHouseCurrentL3, Name: "house_current_3", Title: "House current L3", Unit: "A", Address: 0x000C, ScaleMilli: 1000},
	{Tag: TagHousePowerActiveTotal, Name: "house_power_active_total", Title: "House total active power", Unit: "W", Address: 0x0010, ScaleMilli: 1},
	{Tag: TagHousePowerActiveL1, Name: "house_power_active_1", Title: "House active power L1", Unit: "W", Address: 0x0012, ScaleMilli: 1},
	{Tag: TagHousePowerActiveL2, Name: "house_power_active_2", Title: "House active power L2", Unit: "W", Address: 0x0014, ScaleMilli: 1},
	{Tag: TagHousePowerActiveL3, Name: "house_power_active_3", Title: "House active power L3", Unit: "W", Address: 0x0016, ScaleMilli: 1},
	{Tag: TagHousePowerReactiveTotal, Name: "house_power_reactive_total", Title: "House total reactive power", Unit: "VA", Address: 0x0018, ScaleMilli: 1},
	{Tag: TagHousePowerReactiveL1, Name: "house_power_reactive_1", Title: "House reactive power L1", Unit: "VA", Address: 0x001A, ScaleMilli: 1},
	{Tag: TagHousePowerReactiveL2, Name: "house_power_reactive_2", Title: "House reactive power L2", Unit: "VA", Address: 0x001C, ScaleMilli: 1},
	{Tag: TagHousePowerReactiveL3, Name: "house_power_reactive_3", Title: "House reactive power L3", Unit: "VA", Address: 0x001E, ScaleMilli: 1},
	{Tag: TagHousePowerFactorL1, Name: "house_power_factor_1", Title: "House power factor L1", Unit: "", Address: 0x002A, ScaleMilli: 1000},
	{Tag: TagHousePowerFactorL2, Name: "house_power_factor_2", Title: "House power factor L2", Unit: "", Address: 0x002C, ScaleMilli: 1000},
	{Tag: TagHousePowerFactorL3, Name: "house_power_factor_3", Title: "House power factor L3", Unit: "", Address: 0x002E, ScaleMilli: 1000},
	{Tag: TagHouseFrequency, Name: "house_frequency", Title: "House frequency", Unit: "Hz", Address: 0x0036, ScaleMilli: 1000},
	{Tag: TagHouseEnergyActive, Name: "house_energy_active", Title: "House total active energy", Unit: "kWh", Address: 0x0100, ScaleMilli: 1000},
	{Tag: TagHouseEnergyReactive, Name: "house_energy_reactive", Title: "House total reactive energy", Unit: "kVAh", Address: 0x0400, ScaleMilli: 1000},
}

var AllRegisters = append(append(make([]RegDef, 0, len(GridRegisters)+len(HouseRegisters)), GridRegisters...), HouseRegisters...)
