package commoncfg

import (
	"fmt"
	"net"
	"strconv"
)

var (
	ErrInvalidAddress = fmt.Errorf("invalid address format")
	ErrInvalidPort    = fmt.Errorf("invalid port")
)

type AddressFlagValue struct {
	Host string // "" если не задано
	Port *int   // nil если не задано
}

func (v *AddressFlagValue) isFlagValue() {}

var _ FlagValue = (*AddressFlagValue)(nil)

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
