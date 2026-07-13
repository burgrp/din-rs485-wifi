//go:build tinygo

// Package rs485 implements a minimal Modbus-RTU master over a half-duplex
// RS485 UART. It is board-agnostic: pins, baud rate and timing are supplied
// through Config, so the same client can be reused by different firmware.
package rs485

import (
	"errors"
	"machine"
	"time"
	_ "unsafe"
)

// Default frame timing, applied by New when a Config field is left zero.
const (
	defaultTxSettle    = 2 * time.Millisecond
	defaultRxFirstByte = 120 * time.Millisecond
	defaultRxInterByte = 15 * time.Millisecond
)

// Buffer sizing. The Modbus wire protocol allows up to 125 words per read, but
// this firmware never requests more than 30 (MaxRunCount=15 float32 registers,
// 2 words each). The receive buffers are sized for that application limit to
// save RAM under the leaking GC; raise maxWords if a caller needs larger reads.
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
	errResponseLarge = errors.New("response too large")
	errTimeout       = errors.New("timeout waiting response")
)

// Config describes the RS485 wiring and timing for a Client.
type Config struct {
	// UART is the hardware UART to drive. If nil, machine.DefaultUART is used.
	UART *machine.UART
	// TX and RX are the UART data pins.
	TX machine.Pin
	RX machine.Pin
	// AltFunc is the GPIO alternate-function number that routes the USART to
	// the TX/RX pins (e.g. AF1 for PA2/PA3 = USART1 on PY32F030). The TinyGo
	// py32 UART driver does not route the peripheral to pins itself, so New
	// selects this alternate function on both TX and RX explicitly.
	AltFunc uint8
	// TxEn drives the transceiver's DE/RE direction control (active-high while
	// transmitting).
	TxEn machine.Pin
	// Baud is the serial bit rate (e.g. 9600).
	Baud uint32
	// TxSettle is the guard time held before and after driving the bus. Zero
	// selects defaultTxSettle.
	TxSettle time.Duration
	// RxFirstByte is how long to wait for the first response byte. Zero selects
	// defaultRxFirstByte.
	RxFirstByte time.Duration
	// RxInterByte is the idle gap that ends a response frame. Zero selects
	// defaultRxInterByte.
	RxInterByte time.Duration
}

// Client is a half-duplex RS485 Modbus-RTU master. It is not safe for
// concurrent use; drive it from a single goroutine.
type Client struct {
	uart        *machine.UART
	txEn        machine.Pin
	txSettle    time.Duration
	rxFirstByte time.Duration
	rxInterByte time.Duration
	txFrame     [8]byte
	rxScratch   [64]byte
	rxFrame     [maxFrame]byte
	wordsBuf    [maxWords]uint16
}

// New configures the UART and direction pin and returns a ready Client.
func New(cfg Config) (*Client, error) {
	txEn := cfg.TxEn
	txEn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	txEn.Low()

	u := cfg.UART
	if u == nil {
		u = machine.DefaultUART
	}
	u.Configure(machine.UARTConfig{
		TX:       cfg.TX,
		RX:       cfg.RX,
		BaudRate: cfg.Baud,
	})

	// The py32 UART.Configure does not map USART1 onto the requested pins, so
	// route TX and RX to the peripheral via their alternate function here.
	cfg.TX.Configure(machine.PinConfig{Mode: machine.PinAlternate})
	cfg.TX.SetAltFunc(cfg.AltFunc)
	cfg.RX.Configure(machine.PinConfig{Mode: machine.PinAlternate})
	cfg.RX.SetAltFunc(cfg.AltFunc)

	c := &Client{
		uart:        u,
		txEn:        txEn,
		txSettle:    cfg.TxSettle,
		rxFirstByte: cfg.RxFirstByte,
		rxInterByte: cfg.RxInterByte,
	}
	if c.txSettle == 0 {
		c.txSettle = defaultTxSettle
	}
	if c.rxFirstByte == 0 {
		c.rxFirstByte = defaultRxFirstByte
	}
	if c.rxInterByte == 0 {
		c.rxInterByte = defaultRxInterByte
	}
	return c, nil
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

	c.flushRX()
	c.txEn.High()
	time.Sleep(c.txSettle)
	_, _ = c.uart.Write(c.txFrame[:])
	time.Sleep(c.txSettle)
	c.txEn.Low()

	respLen := 5 + int(qty)*2
	resp, err := c.readFrame(respLen, c.rxFirstByte, c.rxInterByte)
	if err != nil {
		return nil, err
	}

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

func (c *Client) flushRX() {
	for {
		n, _ := c.uart.Read(c.rxScratch[:])
		if n == 0 {
			return
		}
	}
}

func (c *Client) readFrame(expected int, firstTimeout time.Duration, nextTimeout time.Duration) ([]byte, error) {
	if expected > len(c.rxFrame) {
		return nil, errResponseLarge
	}
	buf := c.rxFrame[:expected]
	total := 0
	deadline := monoNanos() + int64(firstTimeout)
	for total < expected {
		n, _ := c.uart.Read(buf[total:])
		if n > 0 {
			total += n
			deadline = monoNanos() + int64(nextTimeout)
			continue
		}
		if monoNanos() >= deadline {
			break
		}
		time.Sleep(250 * time.Microsecond)
	}
	if total < expected {
		return buf[:total], errTimeout
	}
	return buf, nil
}

//go:linkname monoNanos runtime.nanotime
func monoNanos() int64

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
