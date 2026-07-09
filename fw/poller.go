package main

import (
	"encoding/binary"
	"math"

	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// ModbusClient is implemented by the RS485 RTU layer.
type ModbusClient interface {
	ReadInputRegisters(slave uint8, start uint16, qty uint16) ([]uint16, error)
}

// RegisterWriter writes decoded values to the BleRiot-backed register store.
type RegisterWriter interface {
	SetByTag(tag uint8, value int32, valid bool)
}

type regGroup struct {
	first uint16
	last  uint16
	regs  []spec.RegDef
}

// groupRegs mirrors the existing Python grouping heuristic: keep contiguous-ish spans.
func groupRegs(in []spec.RegDef) []regGroup {
	if len(in) == 0 {
		return nil
	}
	out := make([]regGroup, 0, 8)
	cur := regGroup{first: in[0].Address, last: in[0].Address, regs: []spec.RegDef{in[0]}}
	for i := 1; i < len(in); i++ {
		r := in[i]
		if r.Address > cur.last+4 {
			out = append(out, cur)
			cur = regGroup{first: r.Address, last: r.Address, regs: []spec.RegDef{r}}
			continue
		}
		cur.last = r.Address
		cur.regs = append(cur.regs, r)
	}
	out = append(out, cur)
	return out
}

type poller struct {
	client  ModbusClient
	writer  RegisterWriter
	retries int
}

func newPoller(client ModbusClient, writer RegisterWriter, retries int) *poller {
	if retries < 1 {
		retries = 1
	}
	return &poller{client: client, writer: writer, retries: retries}
}

func (p *poller) pollDevice(dev spec.DeviceDef) {
	groups := groupRegs(dev.Registers)
	for _, g := range groups {
		wordCount := 2 * int(g.last-g.first+1)
		var words []uint16
		var err error
		for i := 0; i < p.retries; i++ {
			words, err = p.client.ReadInputRegisters(dev.SlaveAddr, g.first, uint16(wordCount))
			if err == nil {
				break
			}
		}
		if err != nil {
			for _, reg := range g.regs {
				p.writer.SetByTag(reg.Tag, 0, false)
			}
			continue
		}
		for _, reg := range g.regs {
			offset := int(reg.Address - g.first)
			if offset+1 >= len(words) {
				p.writer.SetByTag(reg.Tag, 0, false)
				continue
			}
			f := decodeFloatFromWords(words[offset], words[offset+1])
			if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
				p.writer.SetByTag(reg.Tag, 0, false)
				continue
			}
			v := int32(f * float32(reg.ScaleMilli))
			p.writer.SetByTag(reg.Tag, v, true)
		}
	}
}

// decodeFloatFromWords matches legacy behavior: low-word then high-word (little-endian float32).
func decodeFloatFromWords(lo, hi uint16) float32 {
	var b [4]byte
	binary.LittleEndian.PutUint16(b[0:2], lo)
	binary.LittleEndian.PutUint16(b[2:4], hi)
	u := binary.LittleEndian.Uint32(b[:])
	return math.Float32frombits(u)
}
