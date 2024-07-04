// ------------------------------------------
// Created by (c) 2024 Serge Reinov.
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sergereinov/go-serial/serial"
)

// *** NB by SR @ 2024:
// These are integration tests.
// It is assumed that:
//   1. We have two ports `COMa` and `COMb`.
//   2. And they are connected by a null-modem cable.
//   3. Or a special driver `com0com` is installed.
//
// One can run these tests separately with the following:
//   go test -v ./... -run Timeouts

const (
	_PortA = "COM22"
	_PortB = "COM23"

	// In the windows version there are specified two buffers by 64 bytes size
	// see `setupComm(h, 64, 64)` in `open_windows.go` file.
	_InBufSize  = 64
	_OutBufSize = 64

	// Expected accuracy of timeouts. May depend on go runtime, OS, I/O port driver.
	// Tests may be flaky at small values ​​(less than 10 milliseconds or so).
	defaultTimeoutAccuracy = time.Millisecond * 2
)

type timeoutCase struct {
	name            string
	opFunc          func(wg *sync.WaitGroup, cancelFuncCh chan func(), resultCh chan Result)
	expectOpTime    time.Duration
	timeoutAccuracy time.Duration
	expectValue     int
	cleanupFunc     func() error
}

// We must run all tests sequentially because COM ports cannot be tested in parallel.
func TestTimeouts(t *testing.T) {
	tcAll := []timeoutCase{
		{
			name: "write total timeout",
			opFunc: func(wg *sync.WaitGroup, cancelFuncCh chan func(), resultCh chan Result) {
				defer wg.Done()
				data := make([]byte, _OutBufSize+1) // make sure this overflows the output buffer
				openOpt := newOpenOptions(_PortA)
				timeouts := newWriteTotalTimeout(time.Second * 2)
				writeFunc(cancelFuncCh, resultCh, data, openOpt, timeouts)
			},
			expectOpTime:    time.Second * 2,
			timeoutAccuracy: defaultTimeoutAccuracy,
			expectValue:     -1, // doesn't care
			cleanupFunc:     purgeBothPorts,
		},
		{
			name: "read total timeout",
			opFunc: func(wg *sync.WaitGroup, cancelFuncCh chan func(), resultCh chan Result) {
				defer wg.Done()
				buf := make([]byte, 1) // there are no incoming bytes, so a buffer size of 1 byte is sufficient
				openOpt := newOpenOptions(_PortA)
				timeouts := newReadTotalTimeout(time.Second * 2)
				readFunc(cancelFuncCh, resultCh, buf, openOpt, timeouts)
			},
			expectOpTime:    time.Second * 2,
			timeoutAccuracy: defaultTimeoutAccuracy,
			expectValue:     0, // expects for no incoming bytes
			cleanupFunc:     purgeBothPorts,
		},
		{
			name: "read intercharacter timeout",
			opFunc: func(wg *sync.WaitGroup, cancelFuncCh chan func(), resultCh chan Result) {
				defer wg.Done()

				const dataLen = 10
				dataToSendA := make([]byte, dataLen)
				bufToRecvB := make([]byte, dataLen+1) // make sure len(buf) > len(data) to trigger read intercharacter timeout

				openOptA := newOpenOptions(_PortA)
				timeoutsA := serial.DefaultTimeouts()
				openOptB := newOpenOptions(_PortB)

				const InterCharacterTimeout = 50 // <- this is the target value of expected timeout

				// This doesn't work (at least for Windows)
				openOptB.InterCharacterTimeout = InterCharacterTimeout

				// Fix InterCharacterTimeout by PlatformSpecificOptions
				timeoutsB := newReadIntercharacterTimeout(time.Duration(InterCharacterTimeout) * time.Millisecond)

				// The total expected timeout is calculated as:
				//   transferTime = dataLen * (_startBit(1) + DataBits(8) + StopBits(1) + PARITY_NONE(0)) / BaudRate(9600)
				//   transferTime = 10 * (1 + 8 + 1 + 0) / 9600 = 100 / 9600 = 0,01041667 seconds = 10,41667 milliseconds
				//   expectedTimeout = transferTime + InterCharacterTimeout = 10 + 50 = 60
				// Why tests may flaks:
				//   Virtual ports (including USB virtual COM adapters) sometimes ignore BaudRate and transmit data at maximum speed.
				//   In these cases, transferTime will be a few milliseconds
				// Also
				//   LAN virtual COM adapters can add network overhead (TCP transmission delay, etc.).
				//   Then transferTime will range from a few milliseconds to several hundred milliseconds.

				transferDataFunc(
					cancelFuncCh,
					resultCh,
					dataToSendA,
					bufToRecvB,
					openOptA,
					timeoutsA,
					openOptB,
					timeoutsB,
				)
			},
			expectOpTime:    time.Millisecond * 60,
			timeoutAccuracy: defaultTimeoutAccuracy,
			expectValue:     10, // expects for equals to dataLen
			cleanupFunc:     purgeBothPorts,
		},
	}

	for _, tc := range tcAll {
		testName := fmt.Sprintf("Test %q", tc.name)

		if err := tc.cleanupFunc(); err != nil {
			t.Logf("WARNING: failed to perform initial cleanup, Test case '%s': %s", tc.name, err.Error())
		}

		resultCh := make(chan Result, 1)
		cancelFuncCh := make(chan func())
		var wg sync.WaitGroup

		start := time.Now()
		wg.Add(1)
		go tc.opFunc(&wg, cancelFuncCh, resultCh)

		cancelFunc := <-cancelFuncCh

		func() {
			select {
			case r := <-resultCh:
				if r.err != nil {
					t.Error(testName, "ERROR:", r.err)
					return
				}
				if tc.expectValue >= 0 && r.value != tc.expectValue {
					t.Errorf("%s ERROR: expect value %d got %d", testName, tc.expectValue, r.value)
					return
				}
				since := time.Since(start)
				if since < tc.expectOpTime-tc.timeoutAccuracy {
					t.Errorf("%s ERROR: expect %v +/-%v, got %v", testName, tc.expectOpTime, tc.timeoutAccuracy, since)
					return
				}
			case <-time.After(tc.expectOpTime + tc.timeoutAccuracy):
				t.Errorf("%s ERROR: waited too long (more than expected %v +/-%v)", testName, tc.expectOpTime, tc.timeoutAccuracy)
				cancelFunc()
				return
			}
			t.Log(testName, "OK")
		}()

		wg.Wait()

		if err := tc.cleanupFunc(); err != nil {
			t.Logf("WARNING: failed to perform final cleanup, Test case '%s': %s", tc.name, err.Error())
		}
	}
}

func newOpenOptions(port string) serial.OpenOptions {
	return serial.OpenOptions{
		PortName:              port,
		BaudRate:              9600,
		ParityMode:            serial.PARITY_NONE,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 100,
	}
}

func newWriteTotalTimeout(total time.Duration) serial.Timeouts {
	t := serial.DefaultTimeouts()
	t.WriteTotal = total
	return t
}

func newReadTotalTimeout(total time.Duration) serial.Timeouts {
	t := serial.DefaultTimeouts()
	t.ReadTotal = total
	return t
}

func newReadIntercharacterTimeout(total time.Duration) serial.Timeouts {
	t := serial.DefaultTimeouts()
	t.ReadIntercharacter = total
	return t
}

type Result struct {
	value int
	err   error
}

func readFunc(
	cancelFuncCh chan func(),
	resultCh chan Result,
	buf []byte,
	openOpt serial.OpenOptions,
	timeouts serial.Timeouts,
) {
	var port *serial.Port
	cancelFunc := func() {
		if port != nil {
			port.Close()
		}
	}
	cancelFuncCh <- cancelFunc

	var err error
	if port, err = serial.Open(openOpt); err != nil {
		resultCh <- Result{0, fmt.Errorf("open error: %w", err)}
		return
	}
	defer port.Close()

	var n int
	if n, err = port.ReadWithTimeouts(buf, timeouts); err != nil {
		resultCh <- Result{0, fmt.Errorf("read error: %w", err)}
		return
	}
	resultCh <- Result{n, nil}
}

func writeFunc(
	cancelFuncCh chan func(),
	resultCh chan Result,
	data []byte,
	openOpt serial.OpenOptions,
	timeouts serial.Timeouts,
) {
	var port *serial.Port
	cancelFunc := func() {
		if port != nil {
			port.Close()
		}
	}
	cancelFuncCh <- cancelFunc

	var err error
	if port, err = serial.Open(openOpt); err != nil {
		resultCh <- Result{0, fmt.Errorf("open error: %w", err)}
		return
	}
	defer port.Close()

	var n int
	if n, err = port.WriteWithTimeouts(data, timeouts); err != nil {
		resultCh <- Result{0, fmt.Errorf("write error: %w", err)}
		return
	}
	resultCh <- Result{n, nil}
}

func transferDataFunc(
	cancelFuncCh chan func(),
	resultCh chan Result,
	dataToSendA []byte,
	bufToRecvB []byte,
	openOptA serial.OpenOptions,
	timeoutsA serial.Timeouts,
	openOptB serial.OpenOptions,
	timeoutsB serial.Timeouts,
) {
	var portA, portB *serial.Port
	stopWaitCh := make(chan struct{})
	cancelFunc := func() {
		if portA != nil {
			portA.Close()
		}
		if portB != nil {
			portB.Close()
		}
		close(stopWaitCh)
	}
	cancelFuncCh <- cancelFunc

	var err error
	if portA, err = serial.Open(openOptA); err != nil {
		resultCh <- Result{0, fmt.Errorf("open port A error: %w", err)}
		return
	}
	defer portA.Close()
	if portB, err = serial.Open(openOptB); err != nil {
		resultCh <- Result{0, fmt.Errorf("open port B error: %w", err)}
		return
	}
	defer portB.Close()

	readResultCh := make(chan Result, 1)
	go func() {
		var r Result
		r.value, r.err = portB.ReadWithTimeouts(bufToRecvB, timeoutsB)
		readResultCh <- r
	}()

	var wn int
	if wn, err = portA.WriteWithTimeouts(dataToSendA, timeoutsA); err != nil {
		resultCh <- Result{0, fmt.Errorf("write to port A error: %w", err)}
		return
	}

	var readResult Result
	select {
	case <-stopWaitCh:
		// we are stopping so there is no need to report any results
		return
	case readResult = <-readResultCh:
	}

	if readResult.err != nil {
		resultCh <- Result{0, fmt.Errorf("read error: %w", err)}
		return
	}
	if readResult.value != wn {
		resultCh <- Result{0, fmt.Errorf("len of read data (%d) is not equal to len of written data (%d)", readResult.value, wn)}
		return
	}

	resultCh <- Result{wn, nil}
}

// Purges hardware and software input and output buffers if any.
// Chunks of data could get stuck in virtual adapter buffers at the transport layer.
func purgeBuffers(portName string) error {
	opt := newOpenOptions(portName)
	port, err := serial.Open(opt)
	if err != nil {
		return fmt.Errorf("failed to open port %s: %w", portName, err)
	}
	defer port.Close()

	port.PurgeBuffers(true, true)

	start := time.Now()
	buf := make([]byte, _InBufSize+_OutBufSize)
	for {
		n, err := port.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read from port %s: %w", portName, err)
		}
		if n == 0 {
			break
		}
		if time.Since(start) > time.Millisecond*10 {
			return fmt.Errorf("failed to purge read buffer (%s port), there is too much data there", portName)
		}
	}

	return nil //ok
}

// Purging the buffers of both ports _PortA and _PortB.
func purgeBothPorts() error {
	errA := purgeBuffers(_PortA)
	errB := purgeBuffers(_PortB)
	return errors.Join(errA, errB)
}
