package main

import "errors"

type noopClient struct{}

func newNoopClient() *noopClient {
	return &noopClient{}
}

func (n *noopClient) ReadInputRegisters(slave uint8, start uint16, qty uint16) ([]uint16, error) {
	_ = slave
	_ = start
	_ = qty
	return nil, errors.New("modbus client not wired")
}
