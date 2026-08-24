package main

import "testing"

func TestDeviceMarksOnlyChangesDirty(t *testing.T) {
	device := &Device{}
	const index = uint8(0)
	const mask = uint32(1) << index

	if _, null := device.Read(measurementTags[index]); !null {
		t.Fatal("new register is not null")
	}

	device.update(index, 123, true)
	if got := device.dirty.Swap(0); got != mask {
		t.Fatalf("initial value dirty mask = %#x, want %#x", got, mask)
	}
	device.update(index, 123, true)
	if got := device.dirty.Swap(0); got != 0 {
		t.Fatalf("unchanged value dirty mask = %#x, want 0", got)
	}

	device.update(index, 124, true)
	if got := device.dirty.Swap(0); got != mask {
		t.Fatalf("changed value dirty mask = %#x, want %#x", got, mask)
	}
	device.update(index, 0, false)
	if got := device.dirty.Swap(0); got != mask {
		t.Fatalf("null transition dirty mask = %#x, want %#x", got, mask)
	}
	device.update(index, 0, false)
	if got := device.dirty.Swap(0); got != 0 {
		t.Fatalf("repeated null dirty mask = %#x, want 0", got)
	}

	device.update(index, 124, true)
	if got := device.dirty.Swap(0); got != mask {
		t.Fatalf("recovery dirty mask = %#x, want %#x", got, mask)
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

func TestNotificationStatePublishesOnlyChanges(t *testing.T) {
	var state notificationState
	const index = uint8(0)

	if !state.changed(index, 123, false) {
		t.Fatal("initial value was suppressed")
	}
	if state.changed(index, 123, false) {
		t.Fatal("duplicate value was published")
	}
	if !state.changed(index, 124, false) {
		t.Fatal("changed value was suppressed")
	}
	if !state.changed(index, 0, true) {
		t.Fatal("null transition was suppressed")
	}
	if state.changed(index, 999, true) {
		t.Fatal("duplicate null was published")
	}
	if !state.changed(index, 124, false) {
		t.Fatal("recovery was suppressed")
	}
	if state.changed(index, 124, false) {
		t.Fatal("duplicate recovered value was published")
	}
}
