//go:build tinygo

// Command bms is the DIN RS485 firmware for a single JK BMS. It polls the BMS
// with the read-all-registers command over the NW protocol and publishes each
// decoded field as a wire tag. The field map is compiled in; the hub provisions
// only identity, channel and poll interval.
package main

import (
	"github.com/burgrp/din-rs485-wifi/fw/bmsdev"
	"github.com/burgrp/din-rs485-wifi/fw/nwbms"
	"github.com/burgrp/din-rs485-wifi/fw/rig"
	"github.com/burgrp/din-rs485-wifi/fw/uartline"
)

// bmsBaud is the JK BMS RS485 bit rate.
const bmsBaud = 115200

// bmsSource polls the BMS and writes decoded values to the runtime device. It
// implements rig.Source.
type bmsSource struct {
	client *nwbms.Client
	tags   []uint16
	dev    *rig.Device
	emit   func(tag uint16, value int32)
}

func (s *bmsSource) Tags() []uint16 { return s.tags }

func (s *bmsSource) Poll(dev *rig.Device) {
	s.dev = dev
	p := s.client.ReadAll()
	if p == nil {
		for _, t := range s.tags {
			dev.Set(t, 0, false)
		}
		return
	}
	nwbms.Decode(p, s.emit)
}

func main() {
	line := uartline.New(uartline.Config{
		UART:      rig.UartBus,
		TX:        rig.UartTx,
		RX:        rig.UartRx,
		TxAltFunc: rig.UartTxAF,
		RxAltFunc: rig.UartRxAF,
		TxEn:      rig.TxEn,
		Baud:      bmsBaud,
	})
	s := &bmsSource{
		client: nwbms.New(line),
		tags:   bmsdev.Tags(),
	}
	// The emit closure is built once and reused every poll, so it does not
	// allocate on the hot path.
	s.emit = func(tag uint16, value int32) { s.dev.Set(tag, value, true) }
	rig.Run(s)
}
