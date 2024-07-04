// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

import (
	"syscall"
	"time"
)

// Sets communication timeouts for next IO operations
func (p *serialPort) SetTimeouts(timeouts Timeouts) error {
	if p == nil || p.fd == syscall.Handle(0) || p.fd == syscall.InvalidHandle {
		return ErrInvalidOrNilPort
	}
	cto := defaultWindowsCommTimeouts
	cto.ReadIntervalTimeout = uint32(timeouts.ReadIntercharacter / time.Millisecond)
	cto.ReadTotalTimeoutConstant = uint32(timeouts.ReadTotal / time.Millisecond)
	cto.WriteTotalTimeoutConstant = uint32(timeouts.WriteTotal / time.Millisecond)
	return setCommTimeouts(p.fd, cto)
}
