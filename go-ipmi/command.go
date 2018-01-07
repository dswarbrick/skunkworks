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

type Request struct {
	NetworkFunction uint8
	Command         uint8
	Data            interface{}
}

// AuthCapabilitiesRequest per section 22.13
type AuthCapabilitiesRequest struct {
	ChannelNumber uint8
	PrivLevel     uint8
}

// AuthCapabilitiesResponse per section 22.13
type AuthCapabilitiesResponse struct {
	CompletionCode  uint8
	ChannelNumber   uint8
	AuthTypeSupport uint8
	Status          uint8
	Reserved        uint8
	OEMID           uint16
	OEMAux          uint8
}

// Authentication types
const (
	AuthTypeNone = iota
	AuthTypeMD2
	AuthTypeMD5
	authTypeReserved
	AuthTypePassword
	AuthTypeOEM
)
