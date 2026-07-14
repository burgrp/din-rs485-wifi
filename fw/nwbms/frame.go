package nwbms

// NW frame layout (both directions):
//
//	0x4E 0x57 | len(2, big-endian) | terminal(4) | fn | src | type | payload… |
//	record(4) | 0x68 | crc(4: high 2 unused/0, low 2 = additive sum)
//
// len is the index of the CRC high byte: the checksum covers bytes [0,len) and
// the CRC value sits at frame[len] and frame[len+1]. The additive checksum is a
// plain 16-bit sum of the covered bytes (the top two CRC bytes are unused).
const (
	stx0           = 0x4E
	stx1           = 0x57
	frameSourceGPS = 0x02
	frameTypeRead  = 0x00
	funcReadAll    = 0x06
	addrReadAll    = 0x00
	endSeq         = 0x68

	// readAllLen is the byte length of the read-all request frame.
	readAllLen = 21
)

// chksum is the JK additive checksum: a 16-bit sum of the given bytes.
func chksum(b []byte) uint16 {
	var s uint16
	for i := 0; i < len(b); i++ {
		s += uint16(b[i])
	}
	return s
}

// BuildReadAll writes the read-all-registers request frame into buf (which must
// be at least readAllLen bytes) and returns its length.
func BuildReadAll(buf []byte) int {
	b := buf[:readAllLen]
	for i := range b {
		b[i] = 0
	}
	b[0] = stx0
	b[1] = stx1
	b[2] = 0x00 // data length high
	b[3] = 0x13 // data length low = 19
	// b[4..7] terminal number = 0
	b[8] = funcReadAll
	b[9] = frameSourceGPS
	b[10] = frameTypeRead
	b[11] = addrReadAll
	// b[12..15] record number = 0
	b[16] = endSeq
	crc := chksum(b[:17])
	b[17] = 0
	b[18] = 0
	b[19] = byte(crc >> 8)
	b[20] = byte(crc)
	return readAllLen
}

// Payload validates a received frame and returns its data payload: the bytes
// from index 11 up to the end-of-frame marker (the JK status data-ID stream).
// It returns false if the frame is malformed, truncated, not a read-all reply,
// or fails the checksum.
func Payload(frame []byte) ([]byte, bool) {
	if len(frame) < 13 {
		return nil, false
	}
	if frame[0] != stx0 || frame[1] != stx1 {
		return nil, false
	}
	dataLen := int(frame[2])<<8 | int(frame[3])
	// Need the header (11), at least an empty payload plus the 3-byte
	// end-of-frame marker, and the trailing 2 CRC bytes.
	if dataLen < 14 || len(frame) < dataLen+2 {
		return nil, false
	}
	if frame[8] != funcReadAll {
		return nil, false
	}
	remote := uint16(frame[dataLen])<<8 | uint16(frame[dataLen+1])
	if chksum(frame[:dataLen]) != remote {
		return nil, false
	}
	return frame[11 : dataLen-3], true
}
