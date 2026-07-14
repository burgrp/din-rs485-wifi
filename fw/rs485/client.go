//go:build tinygo

// Package rs485 implements a minimal Modbus-RTU master on top of a shared
// half-duplex RS485 transport (uartline.Line). It builds and validates Modbus
// frames; the wiring, timing and direction control live in uartline, so the
// same physical line can be shared with other protocols.
package rs485

import (
	"errors"

	"github.com/burgrp/din-rs485-wifi/fw/uartline"
)

// Buffer sizing. The Modbus wire protocol allows up to 125 words per read, but
// this firmware never requests more than 30 (15 float32 registers, 2 words
// each). The receive buffers are sized for that application limit to save RAM
// under the leaking GC; raise maxWords if a caller needs larger reads.
const (
	maxWords = 30
	// maxFrame is the largest response frame: slave+func+bytecount + data + CRC.
	maxFrame = 5 + maxWords*2
)

// Sentinel errors. These are package-level so the hot read path never calls
// errors.New, which would allocate on every failed poll (fatal under the
// leaking GC).
var (
	errQtyZero       = errors.New("qty must be > 0")
	errQtyTooLarge   = errors.New("qty exceeds buffer limit")
	errShortResponse = errors.New("short response")
	errSlaveID       = errors.New("unexpected slave id")
	errException     = errors.New("modbus exception response")
	errFuncCode      = errors.New("unexpected function code")
	errByteCount     = errors.New("byte count mismatch")
	errCRC           = errors.New("crc mismatch")
	errTimeout       = errors.New("timeout waiting response")
)

// Client is a half-duplex RS485 Modbus-RTU master over a uartline.Line. It is
// not safe for concurrent use; drive it from a single goroutine.
type Client struct {
	line     *uartline.Line
	txFrame  [8]byte
	rxFrame  [maxFrame]byte
	wordsBuf [maxWords]uint16
}

// New returns a Modbus client that talks over the given line.
func New(line *uartline.Line) *Client {
	return &Client{line: line}
}

// ReadInputRegisters issues Modbus function 0x04 and returns qty 16-bit words.
func (c *Client) ReadInputRegisters(slave uint8, start uint16, qty uint16) ([]uint16, error) {
	if qty == 0 {
		return nil, errQtyZero
	}
	if qty > maxWords {
		return nil, errQtyTooLarge
	}

	c.txFrame[0] = slave
	c.txFrame[1] = 0x04
	c.txFrame[2] = byte(start >> 8)
	c.txFrame[3] = byte(start)
	c.txFrame[4] = byte(qty >> 8)
	c.txFrame[5] = byte(qty)
	crc := modbusCRC16(c.txFrame[:6])
	c.txFrame[6] = byte(crc)
	c.txFrame[7] = byte(crc >> 8)

	c.line.Send(c.txFrame[:])

	respLen := 5 + int(qty)*2
	n := c.line.Receive(c.rxFrame[:respLen], respLen)
	if n == 0 {
		return nil, errTimeout
	}
	resp := c.rxFrame[:n]

	if len(resp) != respLen {
		return nil, errShortResponse
	}
	if resp[0] != slave {
		return nil, errSlaveID
	}
	if resp[1] == 0x84 {
		return nil, errException
	}
	if resp[1] != 0x04 {
		return nil, errFuncCode
	}
	if int(resp[2]) != int(qty)*2 {
		return nil, errByteCount
	}

	gotCRC := uint16(resp[len(resp)-2]) | (uint16(resp[len(resp)-1]) << 8)
	calcCRC := modbusCRC16(resp[:len(resp)-2])
	if gotCRC != calcCRC {
		return nil, errCRC
	}

	words := c.wordsBuf[:qty]
	off := 3
	for i := 0; i < int(qty); i++ {
		words[i] = (uint16(resp[off]) << 8) | uint16(resp[off+1])
		off += 2
	}

	return words, nil
}

func modbusCRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i])
		for b := 0; b < 8; b++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
