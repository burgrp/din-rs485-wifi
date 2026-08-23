package main

import (
	"errors"
	"slices"
	"testing"

	"github.com/burgrp/bleriot-rs485/fw-em/spec"
)

type scriptedTransport struct {
	request  []byte
	response []byte
	err      error
}

func (transport *scriptedTransport) Exchange(request []byte, response []byte) (int, error) {
	transport.request = append(transport.request[:0], request...)
	return copy(response, transport.response), transport.err
}

func TestReadFloat32sUsesManualVector(t *testing.T) {
	transport := &scriptedTransport{
		response: []byte{0x01, 0x04, 0x04, 0x43, 0x6B, 0x58, 0x0E, 0x25, 0xD8},
	}
	client := newModbusClient(transport)

	values, err := client.readFloat32s(1, 0, 1, spec.WordOrderHighFirst)
	if err != nil {
		t.Fatal(err)
	}
	wantRequest := []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0xCB}
	if !slices.Equal(transport.request, wantRequest) {
		t.Fatalf("request = % X, want % X", transport.request, wantRequest)
	}
	if got, want := uint32(values[0]), uint32(0x436B580E); got != want {
		t.Fatalf("raw value = %08X, want %08X", got, want)
	}
}

func TestReadFloat32sSupportsLowWordFirst(t *testing.T) {
	transport := &scriptedTransport{
		response: withCRC([]byte{0x01, 0x04, 0x04, 0x43, 0x6B, 0x58, 0x0E}),
	}
	values, err := newModbusClient(transport).readFloat32s(1, 0, 1, spec.WordOrderLowFirst)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := uint32(values[0]), uint32(0x580E436B); got != want {
		t.Fatalf("raw value = %08X, want %08X", got, want)
	}
}

func TestReadFloat32sRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name     string
		response []byte
		want     error
	}{
		{name: "short", response: []byte{0x01, 0x04, 0x00}, want: errShortResponse},
		{name: "slave", response: withCRC([]byte{0x02, 0x04, 0x04, 0, 0, 0, 0}), want: errSlave},
		{name: "exception", response: withCRC([]byte{0x01, 0x84, 0x02}), want: errException},
		{name: "function", response: withCRC([]byte{0x01, 0x03, 0x04, 0, 0, 0, 0}), want: errFunction},
		{name: "length", response: withCRC([]byte{0x01, 0x04, 0x02, 0, 0}), want: errShortResponse},
		{name: "byte count", response: withCRC([]byte{0x01, 0x04, 0x03, 0, 0, 0, 0}), want: errByteCount},
		{name: "crc", response: []byte{0x01, 0x04, 0x04, 0, 0, 0, 0, 0, 0}, want: errCRC},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := &scriptedTransport{response: test.response}
			_, err := newModbusClient(transport).readFloat32s(1, 0, 1, spec.WordOrderHighFirst)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestReadFloat32sValidatesCountAndTransportError(t *testing.T) {
	client := newModbusClient(&scriptedTransport{})
	if _, err := client.readFloat32s(1, 0, 0, spec.WordOrderHighFirst); err != errCountZero {
		t.Fatalf("zero count error = %v", err)
	}
	if _, err := client.readFloat32s(1, 0, maxFloatsPerRead+1, spec.WordOrderHighFirst); err != errCountTooLarge {
		t.Fatalf("large count error = %v", err)
	}

	want := errors.New("transport failed")
	client = newModbusClient(&scriptedTransport{err: want})
	if _, err := client.readFloat32s(1, 0, 1, spec.WordOrderHighFirst); !errors.Is(err, want) {
		t.Fatalf("transport error = %v, want %v", err, want)
	}
}

func TestModbusCRC16ManualRequest(t *testing.T) {
	if got, want := modbusCRC16([]byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x02}), uint16(0xCB71); got != want {
		t.Fatalf("crc = %04X, want %04X", got, want)
	}
}

func withCRC(frame []byte) []byte {
	crc := modbusCRC16(frame)
	return append(frame, byte(crc), byte(crc>>8))
}
