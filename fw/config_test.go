//go:build !tinygo

package main

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// smallDevice is a second, different Modbus device type used to prove the
// bridge supports mixed device types on one node.
var smallDevice = &spec.Device{
	Name: "test-small",
	Registers: []spec.MeterRegister{
		{Name: "a", Title: "A", Unit: "x", Address: 0x0000},
		{Name: "b", Title: "B", Unit: "x", Address: 0x0002},
	},
}

// pageImageLen marshals a full provisioning page (header + config + crc) exactly
// as the provisioning tool would, and returns its size in bytes.
func pageImageLen(t *testing.T, cfg spec.Config) int {
	t.Helper()
	img, err := config.Marshal([4]byte{}, [16]byte{}, 37, config.SpreadFactorS8, cfg)
	if err != nil {
		t.Fatalf("marshal page: %v", err)
	}
	return len(img)
}

func TestConfigFitsPage(t *testing.T) {
	slaves := []spec.Slave{
		{Group: "grid", Addr: 1, Sel: 1, Device: spec.SinotimerEnergyMeter3P},
		{Group: "house", Addr: 2, Sel: 2, Device: spec.SinotimerEnergyMeter3P},
	}
	cfg := spec.ConfigForSlaves(slaves, 1000)

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, cfg); err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if b.Len() != configBytes {
		t.Fatalf("configBytes const=%d but marshaled=%d", configBytes, b.Len())
	}

	img := pageImageLen(t, cfg)
	if img > int(spec.Chip.PageBytes) {
		t.Fatalf("page image %d bytes exceeds PageBytes %d", img, spec.Chip.PageBytes)
	}
	t.Logf("config = %d bytes, full page image = %d / %d bytes", b.Len(), img, spec.Chip.PageBytes)
}

func TestTwoIdenticalMetersShareRunPool(t *testing.T) {
	slaves := []spec.Slave{
		{Group: "grid", Addr: 1, Sel: 1, Device: spec.SinotimerEnergyMeter3P},
		{Group: "house", Addr: 2, Sel: 2, Device: spec.SinotimerEnergyMeter3P},
	}
	cfg := spec.ConfigForSlaves(slaves, 1000)
	if cfg.Slaves[0].Span != cfg.Slaves[1].Span {
		t.Fatalf("identical devices should share run pool: %#x vs %#x",
			cfg.Slaves[0].Span, cfg.Slaves[1].Span)
	}
	assertRoundTrip(t, cfg)
	assertPlanRegs(t, cfg, 2*len(spec.SinotimerEnergyMeter3P.Registers))
}

func TestMixedDeviceTypes(t *testing.T) {
	slaves := []spec.Slave{
		{Group: "grid", Addr: 1, Sel: 1, Device: spec.SinotimerEnergyMeter3P},
		{Group: "aux", Addr: 3, Sel: 2, Device: smallDevice},
	}
	cfg := spec.ConfigForSlaves(slaves, 1000)
	if cfg.Slaves[0].Span == cfg.Slaves[1].Span {
		t.Fatalf("different devices should not share run pool span")
	}
	if img := pageImageLen(t, cfg); img > int(spec.Chip.PageBytes) {
		t.Fatalf("mixed config page image %d exceeds PageBytes %d", img, spec.Chip.PageBytes)
	}
	assertRoundTrip(t, cfg)
	assertPlanRegs(t, cfg, len(spec.SinotimerEnergyMeter3P.Registers)+len(smallDevice.Registers))

	// The aux slave's tags must use selector 2 and its own addresses.
	typ := spec.TypeForSlaves(slaves)
	wantAux := spec.TagFor(2, 0x0002)
	found := false
	for _, r := range typ.Registers {
		if r.Name == "aux.b" && r.Tag == wantAux {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected register aux.b with tag %#x", wantAux)
	}
}

func assertRoundTrip(t *testing.T, cfg spec.Config) {
	t.Helper()
	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, cfg); err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := decodeConfig(b.Bytes()); got != cfg {
		t.Fatalf("round-trip mismatch\nwant %+v\ngot  %+v", cfg, got)
	}
}

func assertPlanRegs(t *testing.T, cfg spec.Config, want int) {
	t.Helper()
	total := 0
	for _, p := range buildPlan(cfg) {
		total += int(p.count)
	}
	if total != want {
		t.Fatalf("plan covers %d registers, want %d", total, want)
	}
}
