# Pak


Simple Go (golang.org) package to read and write chromium .pak resource files

### Installation

```
go get github.com/disintegration/pak
```

### Example

```go
package main

import (
	"github.com/disintegration/pak"
)

func main() {
	// Read resources from file into memory
	p, err := pak.ReadFile("resources.pak")
	if err != nil {
		panic(err)
	}

	// Add or update resource
	p.Resourses[12345] = []byte(`<!DOCTYPE html><html><h1>Hello from Go!</h1></html>`)

	// Write back to file
	err = pak.WriteFile("resources.pak", p)
	if err != nil {
		panic(err)
	}
}
```
