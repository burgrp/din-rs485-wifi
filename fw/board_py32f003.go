//go:build tinygo

package main

import "github.com/burgrp/din-rs485-wifi/fw/spec"

func boardConfig() spec.NodeConfig {
	return spec.NodeConfig{
		RS485TxPin:   "PA0",
		RS485RxPin:   "PA3",
		RS485TxEnPin: "PA5",
		LedPin:       "PB0",
		DebugPin:     "PA7",
		RFDataPin:    "PA1",
		RFSckPin:     "PA2",
		RFCsPin:      "PA4",
		ReadRetries:  5,
	}
}
