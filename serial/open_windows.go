// ------------------------------------------
// Modified by (c) 2024 Serge Reinov.
//
// Changes:
//   - Removed OVERLAPPED approach.
//   - Added management of OS communication timeouts.
//   - Added function to purge communication buffers.
//
// The old API remains for backward compatibility.
// When using the old API, the old timeouts behavior is retained.
//
// Licensed under the Apache License, Version 2.0.
// Below is the license of the original project.
// ------------------------------------------
// Copyright 2011 Aaron Jacobs. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serial

import (
	"io"
	"syscall"
	"unsafe"
)

type serialPort struct {
	fd syscall.Handle
}

var _ = io.ReadWriteCloser((*serialPort)(nil))

type structDCB struct {
	DCBlength, BaudRate                            uint32
	flags                                          [4]byte
	wReserved, XonLim, XoffLim                     uint16
	ByteSize, Parity, StopBits                     byte
	XonChar, XoffChar, ErrorChar, EofChar, EvtChar byte
	wReserved1                                     uint16
}

func openInternal(options OpenOptions) (*serialPort, error) {
	if len(options.PortName) > 0 && options.PortName[0] != '\\' {
		options.PortName = "\\\\.\\" + options.PortName
	}

	portNamePtr, err := syscall.UTF16PtrFromString(options.PortName)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(
		portNamePtr,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			syscall.CloseHandle(h)
		}
	}()

	if err = setCommState(h, options); err != nil {
		return nil, err
	}
	if err = setupComm(h, 64, 64); err != nil {
		return nil, err
	}
	cto := ctoFromOpenOptions(options)
	if err = setCommTimeouts(h, cto); err != nil {
		return nil, err
	}

	port := new(serialPort)
	port.fd = h

	return port, nil
}

func (p *serialPort) Close() error {
	if p == nil || p.fd == syscall.Handle(0) || p.fd == syscall.InvalidHandle {
		return ErrInvalidOrNilPort
	}
	return syscall.CloseHandle(p.fd)
}

func (p *serialPort) Write(buf []byte) (int, error) {
	if p == nil || p.fd == syscall.Handle(0) || p.fd == syscall.InvalidHandle {
		return 0, ErrInvalidOrNilPort
	}
	var n uint32
	err := syscall.WriteFile(p.fd, buf, &n, nil)
	return int(n), err
}

func (p *serialPort) Read(buf []byte) (int, error) {
	if p == nil || p.fd == syscall.Handle(0) || p.fd == syscall.InvalidHandle {
		return 0, ErrInvalidOrNilPort
	}
	var done uint32
	err := syscall.ReadFile(p.fd, buf, &done, nil)
	return int(done), err
}

var (
	nSetCommState,
	nSetCommTimeouts,
	nSetupComm,
	nPurgeComm uintptr
)

func init() {
	k32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		panic("LoadLibrary " + err.Error())
	}
	defer syscall.FreeLibrary(k32)

	nSetCommState = getProcAddr(k32, "SetCommState")
	nSetCommTimeouts = getProcAddr(k32, "SetCommTimeouts")
	nSetupComm = getProcAddr(k32, "SetupComm")
	nPurgeComm = getProcAddr(k32, "PurgeComm")
}

func getProcAddr(lib syscall.Handle, name string) uintptr {
	addr, err := syscall.GetProcAddress(lib, name)
	if err != nil {
		panic(name + " " + err.Error())
	}
	return addr
}

func setCommState(h syscall.Handle, options OpenOptions) error {
	var params structDCB
	params.DCBlength = uint32(unsafe.Sizeof(params))

	params.flags[0] = 0x01  // fBinary
	params.flags[0] |= 0x10 // Assert DSR

	if options.ParityMode != PARITY_NONE {
		params.flags[0] |= 0x03 // fParity
		params.Parity = byte(options.ParityMode)
	}

	if options.StopBits == 1 {
		params.StopBits = 0
	} else if options.StopBits == 2 {
		params.StopBits = 2
	}

	params.BaudRate = uint32(options.BaudRate)
	params.ByteSize = byte(options.DataBits)

	if options.RTSCTSFlowControl {
		params.flags[0] |= 0x04 // fOutxCtsFlow = 0x1
		params.flags[1] |= 0x20 // fRtsControl = RTS_CONTROL_HANDSHAKE (0x2)
	}

	r, _, err := syscall.SyscallN(nSetCommState, uintptr(h), uintptr(unsafe.Pointer(&params)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func setCommTimeouts(h syscall.Handle, cto WindowsCommTimeouts) error {
	r, _, err := syscall.SyscallN(nSetCommTimeouts, uintptr(h), uintptr(unsafe.Pointer(&cto)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func setupComm(h syscall.Handle, in, out int) error {
	r, _, err := syscall.SyscallN(nSetupComm, uintptr(h), uintptr(in), uintptr(out))
	if r == 0 {
		return err
	}
	return nil
}

func purgeComm(h syscall.Handle, clearRx, clearTx bool) error {
	const (
		PURGE_RXCLEAR = 0x0008
		PURGE_TXCLEAR = 0x0004
	)
	var flags uint32
	if clearRx {
		flags |= PURGE_RXCLEAR
	}
	if clearTx {
		flags |= PURGE_TXCLEAR
	}
	rBool, _, err := syscall.SyscallN(nPurgeComm, uintptr(h), uintptr(flags))
	if rBool == 0 {
		return err
	}
	return nil
}
