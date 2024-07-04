//go:build !windows

// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

// Do nothing on target OS
func (p *serialPort) PurgeBuffers(_, _ bool) error {
	// skip until not implemented
	return nil
}
