package main

import "github.com/burgrp/bleriot-rs485/fw-em/spec"

const maxFloatsPerRead = 8

type modbusError string

func (err modbusError) Error() string { return string(err) }

const (
	errCountZero     modbusError = "modbus: count must be greater than zero"
	errCountTooLarge modbusError = "modbus: count exceeds buffer"
	errShortResponse modbusError = "modbus: short response"
	errSlave         modbusError = "modbus: unexpected slave"
	errFunction      modbusError = "modbus: unexpected function"
	errException     modbusError = "modbus: exception response"
	errByteCount     modbusError = "modbus: byte count mismatch"
	errCRC           modbusError = "modbus: crc mismatch"
)

type modbusTransport interface {
	Exchange(request []byte, response []byte) (int, error)
}

type modbusClient struct {
	transport modbusTransport
	request   [8]byte
	response  [5 + maxFloatsPerRead*4]byte
	values    [maxFloatsPerRead]int32
}

func newModbusClient(transport modbusTransport) *modbusClient {
	return &modbusClient{transport: transport}
}

func (client *modbusClient) readFloat32s(slave uint8, start uint16, count uint8, order spec.WordOrder) ([]int32, error) {
	if count == 0 {
		return nil, errCountZero
	}
	if count > maxFloatsPerRead {
		return nil, errCountTooLarge
	}

	quantity := uint16(count) * 2
	client.request[0] = slave
	client.request[1] = 0x04
	client.request[2] = byte(start >> 8)
	client.request[3] = byte(start)
	client.request[4] = byte(quantity >> 8)
	client.request[5] = byte(quantity)
	crc := modbusCRC16(client.request[:6])
	client.request[6] = byte(crc)
	client.request[7] = byte(crc >> 8)

	expectedLength := 5 + int(count)*4
	n, err := client.transport.Exchange(client.request[:], client.response[:expectedLength])
	if err != nil {
		return nil, err
	}
	if n < 5 || n > expectedLength {
		return nil, errShortResponse
	}
	response := client.response[:n]
	if response[0] != slave {
		return nil, errSlave
	}
	gotCRC := uint16(response[n-2]) | uint16(response[n-1])<<8
	if modbusCRC16(response[:n-2]) != gotCRC {
		return nil, errCRC
	}
	if response[1]&0x80 != 0 {
		return nil, errException
	}
	if response[1] != 0x04 {
		return nil, errFunction
	}
	if n != expectedLength {
		return nil, errShortResponse
	}
	if response[2] != count*4 {
		return nil, errByteCount
	}

	for index := 0; index < int(count); index++ {
		offset := 3 + index*4
		highWord := uint32(response[offset])<<8 | uint32(response[offset+1])
		lowWord := uint32(response[offset+2])<<8 | uint32(response[offset+3])
		if order == spec.WordOrderLowFirst {
			highWord, lowWord = lowWord, highWord
		}
		client.values[index] = int32(highWord<<16 | lowWord)
	}
	return client.values[:count], nil
}

func modbusCRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, value := range data {
		crc ^= uint16(value)
		for bit := uint8(0); bit < 8; bit++ {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
