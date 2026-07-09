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
	d := &runtimeDevice{store: newRegisterStore()}
	for i := range spec.AllRegisters {
		d.store.SetByTag(spec.AllRegisters[i].Tag, 0, false)
	}
	return d
}

func (d *runtimeDevice) bindNode(n *node.Node) {
	d.node = n
}

func (d *runtimeDevice) Read(tag uint16) (value int32, null bool) {
	v, ok := d.store.byTag[uint8(tag)]
	if !ok || !v.valid {
		return 0, true
	}
	return v.value, false
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
