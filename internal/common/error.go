package common

import "errors"

var (
	InvalidArgsErr = errors.New("Invalid Args Error")
	ListErr        = errors.New("List Error")
	AddErr         = errors.New("Add Error")
	RemoveErr      = errors.New("Remove Error")
)
