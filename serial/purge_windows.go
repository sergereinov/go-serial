// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

// Purges input and output buffers.
func (p *serialPort) PurgeBuffers(clearRx, clearTx bool) error {
	return purgeComm(p.fd, clearRx, clearTx)
}
