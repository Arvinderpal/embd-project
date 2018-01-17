package types

import "errors"

var ErrUnknownDriverType = errors.New("unknown driver type")
var ErrUnknownControllerType = errors.New("unknown controller type")
var ErrUnknownAdaptorType = errors.New("unknown adaptor type")
var ErrDriverNotFound = errors.New("driver not found")
var ErrControllerNotFound = errors.New("controller not found")
