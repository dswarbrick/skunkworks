package main

// Command Number Assignments (table G-1)
const (
	CmdGetChannelAuthCapabilities = 0x38
)

// Privilege levels
const (
	PrivLevelUnspecified = iota
	PrivLevelCallback
	PrivLevelUser
	PrivLevelOperator
	PrivLevelAdmin
	PrivLevelOEM
)

// AuthCapabilitiesRequest per section 22.13
type AuthCapabilitiesRequest struct {
	ChannelNumber uint8
	PrivLevel     uint8
}
