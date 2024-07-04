// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

import "time"

type Timeouts struct {
	ReadIntercharacter time.Duration
	ReadTotal          time.Duration
	WriteTotal         time.Duration
}

func DefaultTimeouts() Timeouts {
	return Timeouts{
		ReadIntercharacter: time.Millisecond,
		ReadTotal:          time.Millisecond * 100,
		WriteTotal:         time.Millisecond * 100,
	}
}

func (p *serialPort) ReadWithTimeouts(buf []byte, timeouts Timeouts) (int, error) {
	p.SetTimeouts(timeouts)
	return p.Read(buf)
}

func (p *serialPort) WriteWithTimeouts(buf []byte, timeouts Timeouts) (int, error) {
	p.SetTimeouts(timeouts)
	return p.Write(buf)
}
