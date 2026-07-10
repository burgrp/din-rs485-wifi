package main

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// ModbusClient is implemented by the RS485 RTU layer.
type ModbusClient interface {
	ReadInputRegisters(slave uint8, start uint16, qty uint16) ([]uint16, error)
}

// runPlan is one Modbus read: Count float32 registers starting at BaseAddr on
// slaveAddr, whose results map onto wire tags via sel.
type runPlan struct {
	slaveAddr uint8
	sel       uint8
	baseAddr  uint16
	count     uint8
}

// buildPlan flattens a Config into an ordered list of Modbus reads. Each slave
// polls only its own slice of the shared run pool, so different device types
// can coexist. The order (slaves, then runs, then registers) defines the
// register-table index that runtimeDevice and the poller share.
func buildPlan(cfg spec.Config) []runPlan {
	plan := make([]runPlan, 0, spec.MaxSlaves*spec.MaxRuns)
	for s := range cfg.Slaves {
		sl := cfg.Slaves[s]
		if sl.Addr == 0 || sl.Sel == 0 {
			continue
		}
		start := int(sl.RunStart())
		end := start + int(sl.RunCount())
		for r := start; r < end && r < spec.MaxRuns; r++ {
			run := cfg.Runs[r]
			if run.Count() == 0 {
				continue
			}
			plan = append(plan, runPlan{
				slaveAddr: sl.Addr,
				sel:       sl.Sel,
				baseAddr:  run.Addr(),
				count:     run.Count(),
			})
		}
	}
	return plan
}

// configBytes is the fixed on-wire size of spec.Config (little-endian, no padding).
const configBytes = 2 + spec.MaxRuns*2 + spec.MaxSlaves*3

// decodeConfig parses the provisioning payload without reflection. It mirrors
// encoding/binary.Write over spec.Config exactly.
func decodeConfig(raw []byte) spec.Config {
	var cfg spec.Config
	if len(raw) < configBytes {
		return cfg
	}
	cfg.PollMs = uint16(raw[0]) | uint16(raw[1])<<8
	o := 2
	for i := 0; i < spec.MaxRuns; i++ {
		cfg.Runs[i] = spec.Run(uint16(raw[o]) | uint16(raw[o+1])<<8)
		o += 2
	}
	for i := 0; i < spec.MaxSlaves; i++ {
		cfg.Slaves[i].Addr = raw[o]
		cfg.Slaves[i].Sel = raw[o+1]
		cfg.Slaves[i].Span = raw[o+2]
		o += 3
	}
	return cfg
}

var errShortRead = errors.New("short modbus read")

type poller struct {
	client  ModbusClient
	dev     *runtimeDevice
	plan    []runPlan
	retries int
}

func newPoller(client ModbusClient, dev *runtimeDevice, plan []runPlan, retries int) *poller {
	if retries < 1 {
		retries = 1
	}
	return &poller{client: client, dev: dev, plan: plan, retries: retries}
}

// pollAll reads every planned run and writes decoded values into the device by
// table index (same iteration order used to build the device's tag table).
func (p *poller) pollAll() {
	idx := 0
	for i := range p.plan {
		rp := p.plan[i]
		words, err := p.read(rp.slaveAddr, rp.baseAddr, rp.count)
		for j := uint8(0); j < rp.count; j++ {
			if err != nil {
				p.dev.set(idx, 0, false)
				idx++
				continue
			}
			f := decodeFloatFromWords(words[j*2], words[j*2+1])
			if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
				p.dev.set(idx, 0, false)
			} else {
				p.dev.set(idx, int32(f*float32(spec.WireScaleMilli)), true)
			}
			idx++
		}
	}
}

func (p *poller) read(slave uint8, base uint16, count uint8) ([]uint16, error) {
	qty := uint16(count) * 2
	var words []uint16
	var err error
	for i := 0; i < p.retries; i++ {
		words, err = p.client.ReadInputRegisters(slave, base, qty)
		if err == nil {
			break
		}
	}
	if err == nil && len(words) < int(qty) {
		return nil, errShortRead
	}
	return words, err
}

// decodeFloatFromWords matches legacy behavior: low-word then high-word (little-endian float32).
func decodeFloatFromWords(lo, hi uint16) float32 {
	var b [4]byte
	binary.LittleEndian.PutUint16(b[0:2], lo)
	binary.LittleEndian.PutUint16(b[2:4], hi)
	u := binary.LittleEndian.Uint32(b[:])
	return math.Float32frombits(u)
}
