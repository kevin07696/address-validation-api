package config

import (
	"errors"
	"fmt"
	"strconv"
)

func ParseStringToUint16(s string) (uint16, error) {
	// First convert to int to catch negative numbers
	num, err := ParseInt(s)
	if err != nil {
		return 0, err
	}
	
	// Check if it fits in uint16 range
	if num < 0 || num > 65535 {
		return 0, errors.New("port out of range (0-65535)")
	}

	return uint16(num), nil
}


func ParseStringToUint8(s string) (uint8, error) {
	// First convert to int to catch negative numbers
	num, err := ParseInt(s)
	if err != nil {
		return 0, err
	}
	// Check if it fits in uint16 range
	if num < 0 || num > 255 {
		return 0, errors.New("port out of range (0-255)")
	}

	return uint8(num), nil
}

func ParseInt(s string) (num int, err error) {
	num, err = strconv.Atoi(s)
	if err != nil {
		err = fmt.Errorf("invalid port format: %w", err)
	}
	return
}
