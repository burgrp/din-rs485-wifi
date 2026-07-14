package nwbms

func u16(b []byte, i int) uint16 { return uint16(b[i])<<8 | uint16(b[i+1]) }

func u32(b []byte, i int) uint32 { return uint32(u16(b, i))<<16 | uint32(u16(b, i+2)) }

// temperature decodes a JK temperature word: 0..99 is degrees Celsius, values
// above 99 encode negatives (100 -> -1, 101 -> -2, …). Matches jk_bms.h.
func temperature(v uint16) int32 {
	if v > 99 {
		return 99 - int32(int16(v))
	}
	return int32(v)
}

// current decodes the JK current word into milliamps. Only protocol version 1
// carries a signed reading (top bit = sign, low 15 bits = 10 mA units); other
// versions report zero, as in the reference driver.
func current(v uint16, protocolVersion uint8) int32 {
	if protocolVersion != 0x01 {
		return 0
	}
	mag := int32(v&0x7FFF) * 10 // 10 mA units -> mA
	if v&0x8000 != 0 {
		return mag
	}
	return -mag
}

// Decode walks a read-all status payload and calls emit for each decoded field
// (cell voltages in mV, total voltage in mV, current in mA, temperatures in °C,
// SOC in %, counts and bitmasks as integers). It returns false if the payload
// is too short to hold a full status frame. Decode does not allocate.
func Decode(p []byte, emit func(tag uint16, value int32)) bool {
	if len(p) < 2 {
		return false
	}
	// p[1] is the cell-block byte count (3 bytes per cell). The reference
	// requires the full settings block up to the protocol-version byte.
	if len(p) < int(p[1])+223 {
		return false
	}
	cells := int(p[1]) / 3
	for i := 0; i < cells && i < MaxCells; i++ {
		emit(CellTag(i+1), int32(u16(p, i*3+3))) // millivolts
	}

	off := int(p[1]) + 3
	protocolVersion := p[off+219]

	emit(TagTempMosfet, temperature(u16(p, off+0)))
	emit(TagTempBox, temperature(u16(p, off+3)))
	emit(TagTempBattery, temperature(u16(p, off+6)))
	emit(TagTotalVoltage, int32(u16(p, off+9))*10) // 0.01 V -> mV
	emit(TagCurrent, current(u16(p, off+12), protocolVersion))
	emit(TagSOC, int32(p[off+15]))
	emit(TagCycles, int32(u16(p, off+19)))
	emit(TagCycleCapacity, int32(u32(p, off+22)))
	emit(TagCellCount, int32(u16(p, off+27)))
	emit(TagErrors, int32(u16(p, off+30)))
	emit(TagModes, int32(u16(p, off+33)))
	return true
}
