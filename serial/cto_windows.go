// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

// COMMTIMEOUTS is described in more detail in the documentation:
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-commtimeouts
// About Windows communication resource timeouts:
// https://learn.microsoft.com/en-us/windows/win32/devio/time-outs
type WindowsCommTimeouts struct {
	// The timeout will work when there was an incoming data flow, but it stopped.
	ReadIntervalTimeout uint32

	// Complicated multiplier. It is multiplied by the number of bytes expected when reading.
	// In other words, it is multiplied by the size of the buffer for the read data.
	// More precise formula:
	//   Timeout = (MULTIPLIER * number_of_bytes) + CONSTANT
	//
	// It is better to leave it as zero unless there is a special need.
	ReadTotalTimeoutMultiplier uint32

	// Timeout for when there is no incoming data.
	ReadTotalTimeoutConstant uint32

	// Complicated multiplier. It is multiplied by the number of bytes written.
	// More precise formula:
	//   Timeout = (MULTIPLIER * number_of_bytes) + CONSTANT
	//
	// It is better to leave it as zero unless there is a special need.
	WriteTotalTimeoutMultiplier uint32

	// Timeout to safely handle errors when data cannot be written for some reason.
	// The driver or hardware port may not allow writing while waiting for various permission signals.
	// The wait can be INFINITE until the transmit buffer is cleared enough to continue.
	//
	// It is highly recommended not to leave this timeout at zero.
	WriteTotalTimeoutConstant uint32
}

var defaultWindowsCommTimeouts = WindowsCommTimeouts{
	ReadIntervalTimeout:         1,
	ReadTotalTimeoutMultiplier:  0,
	ReadTotalTimeoutConstant:    100,
	WriteTotalTimeoutMultiplier: 0,
	WriteTotalTimeoutConstant:   100,
}

// This is old code. Moved from the file `open_windows.go` along with the author's comments.
// The code is left unchanged to preserve the default behavior.
func ctoFromOpenOptions(options OpenOptions) WindowsCommTimeouts {
	var timeouts WindowsCommTimeouts
	const MAXDWORD = 1<<32 - 1
	timeoutConstant := uint32(round(float64(options.InterCharacterTimeout) / 100.0))
	readIntervalTimeout := uint32(options.MinimumReadSize)

	if timeoutConstant > 0 && readIntervalTimeout == 0 {
		//Assume we're setting for non blocking IO.
		timeouts.ReadIntervalTimeout = MAXDWORD
		timeouts.ReadTotalTimeoutMultiplier = MAXDWORD
		timeouts.ReadTotalTimeoutConstant = timeoutConstant
	} else if readIntervalTimeout > 0 {
		// Assume we want to block and wait for input.
		timeouts.ReadIntervalTimeout = readIntervalTimeout
		timeouts.ReadTotalTimeoutMultiplier = 1
		timeouts.ReadTotalTimeoutConstant = 1
	} else {
		// No idea what we intended, use defaults
		// default config does what it did before.
		timeouts.ReadIntervalTimeout = MAXDWORD
		timeouts.ReadTotalTimeoutMultiplier = MAXDWORD
		timeouts.ReadTotalTimeoutConstant = MAXDWORD - 1
	}

	/*
			Empirical testing has shown that to have non-blocking IO we need to set:
				ReadTotalTimeoutConstant > 0 and
				ReadTotalTimeoutMultiplier = MAXDWORD and
				ReadIntervalTimeout = MAXDWORD

				The documentation states that ReadIntervalTimeout is set in MS but
				empirical investigation determines that it seems to interpret in units
				of 100ms.

				If InterCharacterTimeout is set at all it seems that the port will block
				indefinitly until a character is received.  Not all circumstances have been
				tested. The input of an expert would be appreciated.

			From http://msdn.microsoft.com/en-us/library/aa363190(v=VS.85).aspx

			 For blocking I/O see below:

			 Remarks:

			 If an application sets ReadIntervalTimeout and
			 ReadTotalTimeoutMultiplier to MAXDWORD and sets
			 ReadTotalTimeoutConstant to a value greater than zero and
			 less than MAXDWORD, one of the following occurs when the
			 ReadFile function is called:

			 If there are any bytes in the input buffer, ReadFile returns
			       immediately with the bytes in the buffer.

			 If there are no bytes in the input buffer, ReadFile waits
		               until a byte arrives and then returns immediately.

			 If no bytes arrive within the time specified by
			       ReadTotalTimeoutConstant, ReadFile times out.
	*/

	return timeouts
}
