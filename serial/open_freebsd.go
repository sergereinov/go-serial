// ------------------------------------------
// Modified by (c) 2024 Serge Reinov.
//   The main code has not been changed.
//   Only compatibility with the new object level has been added.
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
)

type serialPort struct {
	io.ReadWriteCloser
}

var _ = io.ReadWriteCloser((*serialPort)(nil))

func openInternal(_ OpenOptions) (*serialPort, error) {
	return nil, ErrNotImplementedOnOS
}

func (m *serialPort) Read(p []byte) (n int, err error) {
	return 0, ErrNotImplementedOnOS
}

func (m *serialPort) Write(p []byte) (n int, err error) {
	return 0, ErrNotImplementedOnOS
}

func (m *serialPort) Close() error {
	return ErrNotImplementedOnOS
}
