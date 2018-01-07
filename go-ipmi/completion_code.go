package main

import (
	"fmt"
)

type completionCode uint8

// Completion codes per section 5.2
const (
	CommandCompleted = completionCode(0x00)
	ErrNodeBusy      = completionCode(0xc0)
	ErrShortPacket   = completionCode(0xc7)
	ErrInvalidPacket = completionCode(0xcc)
)

// Completion code definitions from table 5-2
var completionCodes = map[completionCode]string{
	CommandCompleted: "Command completed normally",
	ErrNodeBusy:      "Node busy",
	ErrShortPacket:   "Request data length invalid",
	ErrInvalidPacket: "Invalid data field in request",
}

// Error satisfies the error interface so that completionCodes may be returned as errors
func (c completionCode) Error() string {
	if s, ok := completionCodes[c]; ok {
		return s
	}
	return fmt.Sprintf("Completion Code: %X", uint8(c))
}
