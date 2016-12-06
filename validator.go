package main

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	Integer                = &IntegerValidator{}
	NoSpaces               = &NoSpacesValidator{}
	Required               = &RequiredValidator{}
	ValidNonPrivilegedPort = &ValidPortValidator{true}
	ValidPort              = &ValidPortValidator{false}
)

type Validator interface {
	Validate(val string) error
}

type RequiredValidator struct{}

func (v *RequiredValidator) Validate(val string) error {
	if val == "" {
		return fmt.Errorf("Value is required")
	}
	return nil
}

type NoSpacesValidator struct{}

func (v *NoSpacesValidator) Validate(val string) error {
	if strings.Contains(val, " ") {
		return fmt.Errorf("Value cannot contain a space")
	}
	return nil
}

type ValidPortValidator struct {
	onlyNonPrivilegedPorts bool
}

func (v *ValidPortValidator) Validate(val string) error {
	minPort := int64(1)
	maxPort := int64(65535)

	if v.onlyNonPrivilegedPorts {
		minPort = 1024
	}

	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid port, requires %d-%d", minPort, maxPort)
	}

	if i < minPort || i > maxPort {
		return fmt.Errorf("Invalid port, requires %d-%d", minPort, maxPort)
	}

	return nil
}

type IntegerValidator struct {
	onlyNonPrivilegedPorts bool
}

func (v *IntegerValidator) Validate(val string) error {
	_, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}

	return nil
}
