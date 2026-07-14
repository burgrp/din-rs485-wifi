//go:build !tinygo

package nwbms

import "testing"

func TestBuildReadAllChecksum(t *testing.T) {
	var buf [readAllLen]byte
	n := BuildReadAll(buf[:])
	if n != readAllLen {
		t.Fatalf("length = %d, want %d", n, readAllLen)
	}
	if buf[0] != stx0 || buf[1] != stx1 {
		t.Fatalf("bad start bytes %#x %#x", buf[0], buf[1])
	}
	if buf[8] != funcReadAll {
		t.Fatalf("function = %#x, want %#x", buf[8], funcReadAll)
	}
	want := chksum(buf[:17])
	got := uint16(buf[19])<<8 | uint16(buf[20])
	if got != want {
		t.Fatalf("checksum = %#x, want %#x", got, want)
	}
}

// buildStatusFrame wraps a status payload in a valid NW frame with a correct
// additive checksum.
func buildStatusFrame(payload []byte) []byte {
	dataLen := 11 + len(payload) + 3
	frame := make([]byte, dataLen+2)
	frame[0] = stx0
	frame[1] = stx1
	frame[2] = byte(dataLen >> 8)
	frame[3] = byte(dataLen)
	frame[8] = funcReadAll
	copy(frame[11:], payload)
	crc := chksum(frame[:dataLen])
	frame[dataLen] = byte(crc >> 8)
	frame[dataLen+1] = byte(crc)
	return frame
}

func put16(b []byte, i int, v uint16) { b[i] = byte(v >> 8); b[i+1] = byte(v) }

func TestPayloadAndDecode(t *testing.T) {
	// Two-cell status payload: p[1] = cells*3 = 6, then the settings block up
	// to the protocol-version byte at off+219.
	payload := make([]byte, 6+223)
	payload[0] = 0x79
	payload[1] = 6
	payload[2] = 1
	put16(payload, 3, 3821) // cell 1
	payload[5] = 2
	put16(payload, 6, 3834) // cell 2

	off := int(payload[1]) + 3     // 9
	put16(payload, off+0, 0x001D)  // mosfet 29°C
	put16(payload, off+3, 0x001E)  // box 30°C
	put16(payload, off+6, 0x001C)  // battery 28°C
	put16(payload, off+9, 5359)    // total voltage 53.59 V
	put16(payload, off+12, 0x80D0) // current, proto v1, +2080 mA
	payload[off+15] = 15           // SOC 15%
	put16(payload, off+19, 4)      // cycles
	put16(payload, off+27, 14)     // cell count
	put16(payload, off+33, 0x0007) // modes
	payload[off+219] = 0x01        // protocol version

	frame := buildStatusFrame(payload)

	got, ok := Payload(frame)
	if !ok {
		t.Fatal("Payload rejected a valid frame")
	}
	if len(got) != len(payload) {
		t.Fatalf("payload length = %d, want %d", len(got), len(payload))
	}

	values := map[uint16]int32{}
	if !Decode(got, func(tag uint16, value int32) { values[tag] = value }) {
		t.Fatal("Decode rejected a valid payload")
	}

	want := map[uint16]int32{
		CellTag(1):       3821,
		CellTag(2):       3834,
		TagTempMosfet:    29,
		TagTempBox:       30,
		TagTempBattery:   28,
		TagTotalVoltage:  53590,
		TagCurrent:       2080,
		TagSOC:           15,
		TagCycles:        4,
		TagCycleCapacity: 0,
		TagCellCount:     14,
		TagErrors:        0,
		TagModes:         7,
	}
	for tag, w := range want {
		if values[tag] != w {
			t.Errorf("tag %#x = %d, want %d", tag, values[tag], w)
		}
	}
}

func TestPayloadRejectsBadChecksum(t *testing.T) {
	payload := make([]byte, 6+223)
	payload[1] = 6
	frame := buildStatusFrame(payload)
	frame[len(frame)-1] ^= 0xFF // corrupt CRC low byte
	if _, ok := Payload(frame); ok {
		t.Fatal("Payload accepted a frame with a bad checksum")
	}
}

func TestDecodeRejectsShortPayload(t *testing.T) {
	if Decode([]byte{0x79, 6, 0, 0}, func(uint16, int32) {}) {
		t.Fatal("Decode accepted a truncated payload")
	}
}

func TestTemperatureAndCurrent(t *testing.T) {
	if temperature(29) != 29 {
		t.Errorf("temperature(29) = %d, want 29", temperature(29))
	}
	if temperature(100) != -1 {
		t.Errorf("temperature(100) = %d, want -1", temperature(100))
	}
	if current(0x80D0, 0x01) != 2080 {
		t.Errorf("current(0x80D0,1) = %d, want 2080", current(0x80D0, 0x01))
	}
	if current(0x00D0, 0x01) != -2080 {
		t.Errorf("current(0x00D0,1) = %d, want -2080", current(0x00D0, 0x01))
	}
	if current(0x80D0, 0x00) != 0 {
		t.Errorf("current with proto 0 = %d, want 0", current(0x80D0, 0x00))
	}
}
