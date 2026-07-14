//go:build tinygo

package nwbms

import "github.com/burgrp/din-rs485-wifi/fw/uartline"

// rxCap bounds the largest status frame the JK BMS returns (24 cells plus the
// full settings block, ~311 bytes with header and trailer).
const rxCap = 320

// Client polls a JK BMS over a shared RS485 line using the read-all command. It
// is not safe for concurrent use; drive it from a single goroutine.
type Client struct {
	line *uartline.Line
	req  [readAllLen]byte
	rx   [rxCap]byte
}

// New returns a Client that talks over the given line.
func New(line *uartline.Line) *Client {
	c := &Client{line: line}
	BuildReadAll(c.req[:])
	return c
}

// ReadAll issues the read-all-registers command and returns the validated
// status payload (a sub-slice of the client's receive buffer), or nil on any
// framing, timeout or checksum error. The returned slice is valid until the
// next ReadAll call.
func (c *Client) ReadAll() []byte {
	c.line.Send(c.req[:])
	// The status frame is variable length, so read on idle-gap framing alone.
	n := c.line.Receive(c.rx[:], 0)
	if n == 0 {
		return nil
	}
	p, ok := Payload(c.rx[:n])
	if !ok {
		return nil
	}
	return p
}
