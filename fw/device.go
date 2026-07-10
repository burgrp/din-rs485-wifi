package main

import (
	"github.com/burgrp/bleriot/lib/node"
	"github.com/burgrp/din-rs485-wifi/fw/spec"
)

// runtimeDevice implements node.Device. It holds a flat, config-ordered table
// of wire tags and their last-known values. It has no meter knowledge: tags are
// computed purely from the poll plan via spec.TagFor.
type runtimeDevice struct {
	tags   []uint16
	values []int32
	valid  []bool
	node   *node.Node
}

func newRuntimeDevice() *runtimeDevice {
	return &runtimeDevice{}
}

// configure sizes the register table from the poll plan. The tag order must
// match the poller's iteration order so poll results can be written by index.
func (d *runtimeDevice) configure(plan []runPlan) {
	n := 0
	for i := range plan {
		n += int(plan[i].count)
	}
	d.tags = make([]uint16, 0, n)
	d.values = make([]int32, n)
	d.valid = make([]bool, n)
	for i := range plan {
		p := plan[i]
		for j := uint8(0); j < p.count; j++ {
			addr := p.baseAddr + uint16(j)*2
			d.tags = append(d.tags, spec.TagFor(p.sel, addr))
		}
	}
}

func (d *runtimeDevice) bindNode(n *node.Node) {
	d.node = n
}

func (d *runtimeDevice) Read(tag uint16) (value int32, null bool) {
	for i := range d.tags {
		if d.tags[i] == tag {
			if !d.valid[i] {
				return 0, true
			}
			return d.values[i], false
		}
	}
	return 0, true
}

func (d *runtimeDevice) Write(tag uint16, value int32, null bool) {
	// All bridged Modbus registers are read-only from the RF side.
}

// set records a polled value by table index and notifies the hub on change.
func (d *runtimeDevice) set(idx int, value int32, valid bool) {
	if idx < 0 || idx >= len(d.values) {
		return
	}
	if d.values[idx] == value && d.valid[idx] == valid {
		return
	}
	d.values[idx] = value
	d.valid[idx] = valid
	if d.node != nil {
		d.node.Notify(d.tags[idx], value, !valid)
	}
}
