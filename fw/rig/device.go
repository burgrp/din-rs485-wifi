//go:build tinygo

package rig

import "github.com/burgrp/bleriot/lib/node"

// Device implements node.Device with a flat table of wire tags and their
// last-known values. It carries no device semantics: the tag set is supplied
// by the Source, and values are pushed in by Set during each poll.
type Device struct {
	tags   []uint16
	values []int32
	valid  []bool
	node   *node.Node
}

func newDevice(tags []uint16) *Device {
	return &Device{
		tags:   tags,
		values: make([]int32, len(tags)),
		valid:  make([]bool, len(tags)),
	}
}

func (d *Device) bindNode(n *node.Node) { d.node = n }

func (d *Device) Read(tag uint16) (value int32, null bool) {
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

func (d *Device) Write(tag uint16, value int32, null bool) {
	// All bridged registers are read-only from the RF side.
}

// Set records a polled value for tag and notifies the hub on change. Unknown
// tags are ignored, so a Source can only touch tags it declared via Tags.
func (d *Device) Set(tag uint16, value int32, valid bool) {
	for i := range d.tags {
		if d.tags[i] != tag {
			continue
		}
		if d.values[i] == value && d.valid[i] == valid {
			return
		}
		d.values[i] = value
		d.valid[i] = valid
		if d.node != nil {
			d.node.Notify(tag, value, !valid)
		}
		return
	}
}
