//go:build tinygo

package main

import (
	"errors"
	"machine"
	"time"
	_ "unsafe"

	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

type uartRS485Client struct {
	uart        *machine.UART
	txEn        machine.Pin
	txSettle    time.Duration
	rxFirstByte time.Duration
	rxInterByte time.Duration
}

func newUARTRS485Client(cfg spec.NodeConfig, baud uint32) (*uartRS485Client, error) {
	txPin, err := pinByName(cfg.RS485TxPin)
	if err != nil {
		return nil, err
	}
	rxPin, err := pinByName(cfg.RS485RxPin)
	if err != nil {
		return nil, err
	}
	txEnPin, err := pinByName(cfg.RS485TxEnPin)
	if err != nil {
		return nil, err
	}

	txEnPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	txEnPin.Low()

	u := machine.DefaultUART
	u.Configure(machine.UARTConfig{
		TX:       txPin,
		RX:       rxPin,
		BaudRate: baud,
	})

	return &uartRS485Client{
		uart:        u,
		txEn:        txEnPin,
		txSettle:    2 * time.Millisecond,
		rxFirstByte: 120 * time.Millisecond,
		rxInterByte: 15 * time.Millisecond,
	}, nil
}

func (c *uartRS485Client) ReadInputRegisters(slave uint8, start uint16, qty uint16) ([]uint16, error) {
	if qty == 0 {
		return nil, errors.New("qty must be > 0")
	}
	if qty > 125 {
		return nil, errors.New("qty exceeds modbus limit")
	}

	request := make([]byte, 8)
	request[0] = slave
	request[1] = 0x04
	request[2] = byte(start >> 8)
	request[3] = byte(start)
	request[4] = byte(qty >> 8)
	request[5] = byte(qty)
	crc := modbusCRC16(request[:6])
	request[6] = byte(crc)
	request[7] = byte(crc >> 8)

	c.flushRX()
	c.txEn.High()
	time.Sleep(c.txSettle)
	_, _ = c.uart.Write(request)
	time.Sleep(c.txSettle)
	c.txEn.Low()

	respLen := 5 + int(qty)*2
	resp, err := c.readFrame(respLen, c.rxFirstByte, c.rxInterByte)
	if err != nil {
		return nil, err
	}

	if len(resp) != respLen {
		return nil, errors.New("short response")
	}
	if resp[0] != slave {
		return nil, errors.New("unexpected slave id")
	}
	if resp[1] == 0x84 {
		return nil, errors.New("modbus exception response")
	}
	if resp[1] != 0x04 {
		return nil, errors.New("unexpected function code")
	}
	if int(resp[2]) != int(qty)*2 {
		return nil, errors.New("byte count mismatch")
	}

	gotCRC := uint16(resp[len(resp)-2]) | (uint16(resp[len(resp)-1]) << 8)
	calcCRC := modbusCRC16(resp[:len(resp)-2])
	if gotCRC != calcCRC {
		return nil, errors.New("crc mismatch")
	}

	words := make([]uint16, qty)
	off := 3
	for i := 0; i < int(qty); i++ {
		words[i] = (uint16(resp[off]) << 8) | uint16(resp[off+1])
		off += 2
	}

	return words, nil
}

func (c *uartRS485Client) flushRX() {
	buf := make([]byte, 64)
	for {
		n, _ := c.uart.Read(buf)
		if n == 0 {
			return
		}
	}
}

func (c *uartRS485Client) readFrame(expected int, firstTimeout time.Duration, nextTimeout time.Duration) ([]byte, error) {
	buf := make([]byte, expected)
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
		return buf[:total], errors.New("timeout waiting response")
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

func pinByName(name string) (machine.Pin, error) {
	switch name {
	case "PA0":
		return machine.PA0, nil
	case "PA1":
		return machine.PA1, nil
	case "PA2":
		return machine.PA2, nil
	case "PA3":
		return machine.PA3, nil
	case "PA4":
		return machine.PA4, nil
	case "PA5":
		return machine.PA5, nil
	case "PA7":
		return machine.PA7, nil
	case "PB0":
		return machine.PB0, nil
	default:
		return 0, errors.New("unknown pin: " + name)
	}
}
