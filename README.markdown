Fork
====
This is a fork of the [go-serial](https://github.com/jacobsa/go-serial) library. The original readme is below the double separation line.

The main changes are related to the lack of timeout settings in the original project.
Why did it seem important to me to improve the library?

The nature of all I/O operations (and also serial I/O) is a waiting.
There is no reason for continuously checking every couple of milliseconds to see if there is any data received?
Having timeout settings helps free up the CPU and simplify tasks. And it will be easier not to miss the moment when the time comes to process the data.
With a trade-off in the form of a slightly slower reaction to unplanned events such as exiting the program.

Changes made to the Windows version of the library:
 - Removed OVERLAPPED approach.
 - Added management of OS communication timeouts.
 - Added function to purge communication buffers.

The old API remains unchanged for backward compatibility.
When using the old API, the old timeout behavior is retained.

`serial.Open` now returns a structure instead of an interface.
```go
func Open(options OpenOptions) (*Port, error)
```
However, `serial.Port` still implements the `io.ReadWriteCloser`, so it can be passed in arguments by the interface as before.

Several additional methods have been added to `serial.Port`.
They can be described by the following interface:
```go
type Timeouts struct {
	ReadIntercharacter time.Duration
	ReadTotal          time.Duration
	WriteTotal         time.Duration
}

type Timeouter interface {
  // Sets communication timeouts for all subsequent Read() and Write() operations.
  SetTimeouts(timeouts Timeouts) error
  // Sets communication timeouts and reads data within the timeout.
  ReadWithTimeouts(buf []byte, timeouts Timeouts) (int, error)
  // Sets communication timeouts and writes data within the timeout.
  WriteWithTimeouts(buf []byte, timeouts Timeouts) (int, error)
}
```

Added function to purge communication buffers:
```go
PurgeBuffers(clearRx, clearTx bool) error
```

Added neat integration tests for timeouts in the `timeouts_test.go` file.
It describes the expected behavior of ports after setting timeouts.

All improvements were made only for the Windows version of the library. Versions for Linux and other OSes have retained the same behavior as before.

SR.

---
---

go-serial
=========

This is a package that allows you to read from and write to serial ports in Go.


OS support
----------

Currently this package works only on OS X, Linux and Windows. It could probably be ported
to other Unix-like platforms simply by updating a few constants; get in touch if
you are interested in helping and have hardware to test with.


Installation
------------

Simply use `go get`:

    go get github.com/jacobsa/go-serial/serial

To update later:

    go get -u github.com/jacobsa/go-serial/serial


Use
---

Set up a `serial.OpenOptions` struct, then call `serial.Open`. For example:

````go
    import "fmt"
    import "log"
    import "github.com/jacobsa/go-serial/serial"

    ...

    // Set up options.
    options := serial.OpenOptions{
      PortName: "/dev/tty.usbserial-A8008HlV",
      BaudRate: 19200,
      DataBits: 8,
      StopBits: 1,
      MinimumReadSize: 4,
    }

    // Open the port.
    port, err := serial.Open(options)
    if err != nil {
      log.Fatalf("serial.Open: %v", err)
    }

    // Make sure to close it later.
    defer port.Close()

    // Write 4 bytes to the port.
    b := []byte{0x00, 0x01, 0x02, 0x03}
    n, err := port.Write(b)
    if err != nil {
      log.Fatalf("port.Write: %v", err)
    }

    fmt.Println("Wrote", n, "bytes.")
````

See the documentation for the `OpenOptions` struct in `serial.go` for more
information on the supported options.
