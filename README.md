**Memory byte buffer** implementation that plays nicely with the excellent `io` packages

100% test coverage -- if you'd like to contribute tests(bytes\_test.go), please do, there can never be enough

### get

```shell
$ go get github.com/dorind/mbytes
```

### implemented interfaces
- io.Seeker
- io.Reader
- io.ReaderAt
- io.Writer
- io.WriteAt
- io.ByteReader
- io.ByteWriter

### simple usage example

```go
package main

import (
	"fmt"

	"github.com/dorind/mbytes"
)

func main() {
	// create a buffer
	b := mbytes.NewByteBuffer(0)

	// our buffer
	soriginal := "abracadabra"

	// store length for reuse later
	l := len(soriginal)

	// write string to memory buffer
	n, err := b.Write([]byte(soriginal))

	// check for error
	if err != nil {
		panic("WRITE: " + err.Error())
	}

	// validate number of written bytes
	if n != l {
		panic(fmt.Sprintf("Something horribly wrong, attempted to write %v bytes, wrote %v", l, n))
	}

	// move position to the beginning in order to read and validate what we wrote
	pos, err := b.SeekToStart()
	if err != nil {
		panic("SEEK: " + err.Error())
	}

	if pos != 0 {
		panic(fmt.Sprintf("Something horribly wrong, attempted to seek to zero, reported position %v", pos))
	}

	// create a buffer for read
	rbuff := make([]byte, l)
	n, err = b.Read(rbuff)
	if err != nil {
		panic("READ: " + err.Error())
	}

	if n != l {
		panic(fmt.Sprintf("Something horribly wrong, attempted to read %v bytes, read %v", l, n))
	}

	fmt.Println("ORIG", soriginal)
	fmt.Println("READ", string(rbuff))
	fmt.Println("VALID?", soriginal == string(rbuff))
}
```
