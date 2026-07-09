package main

import (
	"github.com/burgrp/bleriot/lib/node"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

type runtimeDevice struct {
	store *registerStore
	node  *node.Node
}

func newRuntimeDevice() *runtimeDevice {
	return &runtimeDevice{store: newRegisterStore(spec.MaxTagForSlots(2))}
}

func (d *runtimeDevice) bindNode(n *node.Node) {
	d.node = n
}

func (d *runtimeDevice) Read(tag uint16) (value int32, null bool) {
	v, valid, ok := d.store.GetByTag(uint8(tag))
	if !ok || !valid {
		return 0, true
	}
	return v, false
}

func (d *runtimeDevice) Write(tag uint16, value int32, null bool) {
	_ = tag
	_ = value
	_ = null
	// All bridged Modbus registers are read-only from the RF side.
}

func (d *runtimeDevice) SetByTag(tag uint8, value int32, valid bool) {
	changed := d.store.SetByTag(tag, value, valid)
	if !changed || d.node == nil {
		return
	}
	d.node.Notify(uint16(tag), value, !valid)
}
