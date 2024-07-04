// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

import "errors"

var (
	ErrNotImplementedOnOS = errors.New("not implemented on this OS")
	ErrInvalidOrNilPort   = errors.New("invalid port")
)
