package types

import "errors"

var ErrUnknownDriverType = errors.New("unknown driver type")
var ErrUnknownAdaptorType = errors.New("unknown adaptor type")
var ErrDriverNotFound = errors.New("driver not found")
