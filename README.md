## in-memdb

The `inmemdb` package provides a simple in-memory database implementation in Go. It can be used as a package in other Go projects to store key-value pairs in memory. The package is safe to use concurrently with multiple goroutines.

### Installation

To use this package in your Go project, you can install it using `go get`:

```bash
go get github.com/knbr13/inmemdb
```

### Usage

```go
package main

import (
	"fmt"

	inmemdb "github.com/knbr13/in-memdb"
)

func main() {
	// Create a new in-memory database
	db := inmemdb.New[string, string]()

	// Set key-value pairs
	db.Set("key1", "value1")
	db.Set("key2", "value2")

	// Get values by key
	value1, ok1 := db.Get("key1")
	value2, ok2 := db.Get("key2")

	if ok1 {
		fmt.Println("Value for key1:", value1)
	}

	if ok2 {
		fmt.Println("Value for key2:", value2)
	}

	// Delete a key
	db.Delete("key1")

	// Transfer data to another database
	anotherDB := inmemdb.New[string, string]()
	db.TransferTo(anotherDB)

	// Copy data to another database
	copyDB := inmemdb.New[string, string]()
	anotherDB.CopyTo(copyDB)

	// Retrieve keys
	keys := anotherDB.Keys()
	fmt.Println("Keys in anotherDB:", keys)

	keys = copyDB.Keys()
	fmt.Println("Keys in copyDB:", keys)
}
```

### Contributing

Contributions are welcome! 
If you find any bugs or have suggestions for improvements, please open an issue or submit a pull request on GitHub.