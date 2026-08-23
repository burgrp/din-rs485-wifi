//go:build tinygo

package main

import (
	"device/py32"
	"machine"
	"time"
	_ "unsafe"

	"github.com/burgrp/bleriot-rs485/fw-em/spec"
)

const (
	pinRS485TX   = machine.PB6
	pinRS485RX   = machine.PB7
	pinRS485TXEN = machine.PF4
)

type rs485Error string

func (err rs485Error) Error() string { return string(err) }

const (
	errUARTWrite       rs485Error = "rs485: incomplete UART write"
	errResponseTimeout rs485Error = "rs485: response timeout"
)

type rs485Transport struct {
	uart             *machine.UART
	txEnable         machine.Pin
	silentInterval   time.Duration
	firstByteTimeout time.Duration
	interByteTimeout time.Duration
}

func newRS485Transport(config spec.Config) (*rs485Transport, error) {
	txEnable := pinRS485TXEN
	txEnable.Configure(machine.PinConfig{Mode: machine.PinOutput})
	txEnable.Low()

	uart := machine.DefaultUART
	if err := uart.Configure(machine.UARTConfig{BaudRate: config.Baud, TX: pinRS485TX, RX: pinRS485RX}); err != nil {
		return nil, err
	}
	configureParity(uart, config.Parity)

	frameBits := int64(11)
	if config.Parity == spec.ParityNone {
		frameBits = 10
	}
	characterNanos := (int64(time.Second)*frameBits + int64(config.Baud) - 1) / int64(config.Baud)
	silentInterval := time.Duration((characterNanos*35 + 9) / 10)
	return &rs485Transport{
		uart:             uart,
		txEnable:         txEnable,
		silentInterval:   silentInterval,
		firstByteTimeout: 120 * time.Millisecond,
		interByteTimeout: silentInterval,
	}, nil
}

func configureParity(uart *machine.UART, parity spec.Parity) {
	const frameMask = py32.USART_CR1_M | py32.USART_CR1_PCE | py32.USART_CR1_PS
	control := uart.Bus.CR1.Get()
	uart.Bus.CR1.Set(control &^ py32.USART_CR1_UE)
	control &^= frameMask
	switch parity {
	case spec.ParityEven:
		control |= py32.USART_CR1_M | py32.USART_CR1_PCE
	case spec.ParityOdd:
		control |= py32.USART_CR1_M | py32.USART_CR1_PCE | py32.USART_CR1_PS
	}
	uart.Bus.CR1.Set(control)
}

func (transport *rs485Transport) Exchange(request []byte, response []byte) (int, error) {
	transport.flushReceiveBuffer()
	time.Sleep(transport.silentInterval)
	transport.txEnable.High()
	time.Sleep(10 * time.Microsecond)
	n, err := transport.uart.Write(request)
	transport.txEnable.Low()
	if err != nil {
		return 0, err
	}
	if n != len(request) {
		return 0, errUARTWrite
	}

	total := 0
	deadline := monotonicNanos() + int64(transport.firstByteTimeout)
	for total < len(response) {
		n, _ = transport.uart.Read(response[total:])
		if n > 0 {
			total += n
			deadline = monotonicNanos() + int64(transport.interByteTimeout)
			continue
		}
		if monotonicNanos() >= deadline {
			if total == 0 {
				return 0, errResponseTimeout
			}
			return total, nil
		}
		time.Sleep(250 * time.Microsecond)
	}
	return total, nil
}

func (transport *rs485Transport) flushReceiveBuffer() {
	for transport.uart.Buffered() > 0 {
		_, _ = transport.uart.ReadByte()
	}
}

//go:linkname monotonicNanos runtime.nanotime
func monotonicNanos() int64
