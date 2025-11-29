package commoncfg

import (
	"fmt"
	"net"
	"strconv"
)

var (
	// ErrInvalidAddress indicates that the address flag could not be parsed.
	ErrInvalidAddress = fmt.Errorf("invalid address format")
	// ErrInvalidPort indicates that the port portion of the flag is invalid.
	ErrInvalidPort = fmt.Errorf("invalid port")
)

// AddressFlagValue stores the parsed host and optional port from a CLI flag.
type AddressFlagValue struct {
	Host string
	Port *int
}

var _ FlagValue = (*AddressFlagValue)(nil)

// ParseAddressFlag parses host:port values provided to command-line flags.
func ParseAddressFlag(value string, present bool) (AddressFlagValue, error) {
	if !present {
		return AddressFlagValue{}, nil
	}
	host, portStr, err := net.SplitHostPort(value)
	if err != nil {
		return AddressFlagValue{}, ErrInvalidAddress
	}
	p, err := strconv.Atoi(portStr)
	if err != nil || p <= 0 {
		return AddressFlagValue{}, ErrInvalidPort
	}
	return AddressFlagValue{Host: host, Port: &p}, nil
}
