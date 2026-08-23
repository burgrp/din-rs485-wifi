package spec

import (
	"math"
	"slices"
	"testing"
)

func TestTypeMatchesDeployedRegisterContract(t *testing.T) {
	deviceType := Type()
	if err := deviceType.Validate(); err != nil {
		t.Fatal(err)
	}
	if got, want := len(deviceType.Registers), 20; got != want {
		t.Fatalf("register count = %d, want %d", got, want)
	}

	wantNames := []string{
		"voltage.1", "voltage.2", "voltage.3",
		"current.1", "current.2", "current.3",
		"power.active.total", "power.active.1", "power.active.2", "power.active.3",
		"power.reactive.total", "power.reactive.1", "power.reactive.2", "power.reactive.3",
		"power.factor.1", "power.factor.2", "power.factor.3",
		"frequency", "energy.active", "energy.reactive",
	}
	wantTags := []uint16{
		0x0001, 0x0003, 0x0005,
		0x0009, 0x000B, 0x000D,
		0x0011, 0x0013, 0x0015, 0x0017,
		0x0019, 0x001B, 0x001D, 0x001F,
		0x002B, 0x002D, 0x002F,
		0x0037, 0x0101, 0x0401,
	}
	wantTitles := []string{
		"Voltage L1", "Voltage L2", "Voltage L3",
		"Current L1", "Current L2", "Current L3",
		"Total active power", "Active power L1", "Active power L2", "Active power L3",
		"Total reactive power", "Reactive power L1", "Reactive power L2", "Reactive power L3",
		"Power factor L1", "Power factor L2", "Power factor L3",
		"Frequency", "Total active energy", "Total reactive energy",
	}
	wantUnits := []string{
		"V", "V", "V", "A", "A", "A",
		"W", "W", "W", "W", "VA", "VA", "VA", "VA",
		"", "", "", "Hz", "kWh", "kVAh",
	}
	for index, register := range deviceType.Registers {
		if register.Tag != wantTags[index] {
			t.Errorf("register %q tag = %d, want %d", register.Name, register.Tag, wantTags[index])
		}
		if register.Name != wantNames[index] {
			t.Errorf("register %d name = %q, want %q", index+1, register.Name, wantNames[index])
		}
		if !register.ReadOnly {
			t.Errorf("register %q is writable", register.Name)
		}
		if register.Type != "float" {
			t.Errorf("register %q type = %q, want float", register.Name, register.Type)
		}
		if register.Metadata["title"] != wantTitles[index] {
			t.Errorf("register %q title = %q, want %q", register.Name, register.Metadata["title"], wantTitles[index])
		}
		if register.Metadata["unit"] != wantUnits[index] {
			t.Errorf("register %q unit = %q, want %q", register.Name, register.Metadata["unit"], wantUnits[index])
		}
		if register.Conversion.Decode == nil || register.Conversion.Encode != nil {
			t.Errorf("register %q does not have a decode-only conversion", register.Name)
		}
	}
}

func TestPollRunsCoverEveryRegister(t *testing.T) {
	wantRuns := []Run{
		{Base: 0x0000, First: 0, Count: 3},
		{Base: 0x0008, First: 3, Count: 3},
		{Base: 0x0010, First: 6, Count: 8},
		{Base: 0x002A, First: 14, Count: 3},
		{Base: 0x0036, First: 17, Count: 1},
		{Base: 0x0100, First: 18, Count: 1},
		{Base: 0x0400, First: 19, Count: 1},
	}
	if !slices.Equal(PollRuns[:], wantRuns) {
		t.Fatalf("poll runs = %#v, want %#v", PollRuns, wantRuns)
	}

	covered := make([]bool, len(Type().Registers))
	for _, run := range PollRuns {
		for offset := uint8(0); offset < run.Count; offset++ {
			index := int(run.First + offset)
			if index >= len(covered) {
				t.Fatalf("run %#v covers out-of-range register %d", run, index)
			}
			if covered[index] {
				t.Fatalf("register %d is covered more than once", index)
			}
			covered[index] = true
			wantTag := run.Base + uint16(offset)*2 + 1
			if got := Type().Registers[index].Tag; got != wantTag {
				t.Errorf("register %d tag = %#04x, run implies %#04x", index, got, wantTag)
			}
		}
	}
	for index, ok := range covered {
		if !ok {
			t.Errorf("register %d is not covered by a poll run", index)
		}
	}
}

func TestDecodeFloat32(t *testing.T) {
	decoded, err := decodeFloat32(int32(0x436b580e))
	if err != nil {
		t.Fatal(err)
	}
	want := float64(math.Float32frombits(0x436b580e))
	if decoded != want {
		t.Fatalf("decoded value = %v, want %v", decoded, want)
	}
}

func TestDecodeFloat32RejectsNonFiniteValues(t *testing.T) {
	for _, raw := range []uint32{0x7f800000, 0xff800000, 0x7fc00000} {
		if _, err := decodeFloat32(int32(raw)); err == nil {
			t.Errorf("decodeFloat32(%08x) succeeded", raw)
		}
	}
}
