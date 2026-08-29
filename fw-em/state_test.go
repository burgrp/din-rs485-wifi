package main

import "testing"

func TestDeviceReadTracksUpdates(t *testing.T) {
	device := &Device{}
	const index = uint8(0)
	tag := measurementTags[index]

	if _, null := device.Read(tag); !null {
		t.Fatal("new register is not null")
	}

	device.update(index, 123, true)
	if value, null := device.Read(tag); value != 123 || null {
		t.Fatalf("initial update = (%d, %t), want (123, false)", value, null)
	}

	device.update(index, 124, true)
	if value, null := device.Read(tag); value != 124 || null {
		t.Fatalf("changed update = (%d, %t), want (124, false)", value, null)
	}
	device.update(index, 0, false)
	if _, null := device.Read(tag); !null {
		t.Fatal("invalidated register is not null")
	}

	device.update(index, 124, true)
	if value, null := device.Read(tag); value != 124 || null {
		t.Fatalf("recovered update = (%d, %t), want (124, false)", value, null)
	}
}

func TestMeasurementFreshnessExpiresOnce(t *testing.T) {
	const timeout = int64(15_000)
	var freshness measurementFreshness

	if freshness.expire(timeout, timeout) {
		t.Fatal("measurement expired before its first value")
	}
	freshness.record(100)
	if freshness.expire(100+timeout-1, timeout) {
		t.Fatal("measurement expired before timeout")
	}
	if !freshness.expire(100+timeout, timeout) {
		t.Fatal("measurement did not expire at timeout")
	}
	if freshness.expire(100+timeout+1, timeout) {
		t.Fatal("measurement expired more than once")
	}

	freshness.record(200 + timeout)
	if !freshness.expire(200+2*timeout, timeout) {
		t.Fatal("recovered measurement did not expire again")
	}
}
