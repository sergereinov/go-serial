// ------------------------------------------
// Modified by (c) 2024 Serge Reinov.
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

// Package serial provides routines for interacting with serial ports.
// Currently it supports only OS X; see the readme file for details.

package serial

import (
	"io"
)

type Port struct {
	*serialPort
}

var _ = io.ReadWriteCloser((*Port)(nil))

// Open creates a `serial.Port` based on the supplied options struct.
// It implements io.ReadWriteCloser interface.
func Open(options OpenOptions) (*Port, error) {
	// Redirect to the OS-specific function.
	port, err := openInternal(options)
	if err != nil {
		return nil, err
	}
	return &Port{port}, nil
}
