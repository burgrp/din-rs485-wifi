//go:build tinygo

// Command meter is the DIN RS485 firmware for a single Sinotimer 3-phase energy
// meter. It polls the meter's Modbus input registers and publishes each as a
// wire tag equal to its Modbus word address. The register map is compiled in;
// the hub provisions only identity, channel and poll interval.
package main

import (
	"encoding/binary"
	"math"

	"github.com/burgrp/din-rs485-wifi/fw/meterdev"
	"github.com/burgrp/din-rs485-wifi/fw/rig"
	"github.com/burgrp/din-rs485-wifi/fw/rs485"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
	"github.com/burgrp/din-rs485-wifi/fw/uartline"
)

const (
	slaveAddr   = 1
	modbusBaud  = 9600
	readRetries = 5
)

// meterSource polls the meter over Modbus and writes decoded values to the
// runtime device. It implements rig.Source.
type meterSource struct {
	client  *rs485.Client
	runs    []meterdev.Run
	retries int
}

func (m *meterSource) Tags() []uint16 {
	tags := make([]uint16, len(meterdev.Registers))
	for i := range meterdev.Registers {
		tags[i] = meterdev.Tag(meterdev.Registers[i].Address)
	}
	return tags
}

func (m *meterSource) Poll(dev *rig.Device) {
	for _, r := range m.runs {
		words, err := m.read(r.Base, r.Count)
		for j := uint8(0); j < r.Count; j++ {
			tag := meterdev.Tag(r.Base + uint16(j)*2)
			if err != nil {
				dev.Set(tag, 0, false)
				continue
			}
			f := decodeFloatFromWords(words[j*2], words[j*2+1])
			if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
				dev.Set(tag, 0, false)
			} else {
				dev.Set(tag, int32(f*float32(spec.WireScaleMilli)), true)
			}
		}
	}
}

func (m *meterSource) read(base uint16, count uint8) ([]uint16, error) {
	qty := uint16(count) * 2
	var words []uint16
	var err error
	for i := 0; i < m.retries; i++ {
		words, err = m.client.ReadInputRegisters(slaveAddr, base, qty)
		if err == nil {
			break
		}
	}
	return words, err
}

// decodeFloatFromWords reassembles a float32 from two Modbus words, low word
// first (little-endian byte order within and across words).
func decodeFloatFromWords(lo, hi uint16) float32 {
	var b [4]byte
	binary.LittleEndian.PutUint16(b[0:2], lo)
	binary.LittleEndian.PutUint16(b[2:4], hi)
	return math.Float32frombits(binary.LittleEndian.Uint32(b[:]))
}

func main() {
	line := uartline.New(uartline.Config{
		UART:    rig.UartBus,
		TX:      rig.UartTx,
		RX:      rig.UartRx,
		AltFunc: rig.UartAF,
		TxEn:    rig.TxEn,
		Baud:    modbusBaud,
	})
	rig.Run(&meterSource{
		client:  rs485.New(line),
		runs:    meterdev.Runs(),
		retries: readRetries,
	})
}
