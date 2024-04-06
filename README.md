## incache

The `incache` package provides a simple cache implementation in Go. It can be used as a package in other Go projects to store key-value pairs in memory. The package is safe to use concurrently with multiple goroutines.

### Installation

To use this package in your Go project, you can install it using `go get`:

```bash
go get github.com/knbr13/incache
```

### Usage

```go
package main

import (
	"fmt"
	"time"

	incache "github.com/knbr13/incache"
)

func main() {
	// Create a new Cache Builder
	cb := incache.New[string, string](3) // 3 is the maximum number of key-value pairs the cache can hold

	// Build the Cache
	db := cb.Build()

	fmt.Println("keys:", db.Keys())

	// Set key-value pairs
	db.Set("key1", "value1")
	db.Set("key2", "value2")

	db.SetWithTimeout("key3", "value3", time.Second)

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
	anotherCB := incache.New[string, string](2)
	anotherDB := anotherCB.Build()
	db.TransferTo(anotherDB)

	// Copy data to another database
	copyCB := incache.New[string, string](2)
	copyDB := copyCB.Build()
	anotherDB.CopyTo(copyDB)

	// Retrieve keys
	keys := anotherDB.Keys()
	fmt.Println("Keys in anotherDB:", keys)

	keys = copyDB.Keys()
	fmt.Println("Keys in copyDB:", keys)

	time.Sleep(time.Second)
	value3, ok3 := copyDB.Get("key3")
	fmt.Printf("ok = %v, value = %v\n", ok3, value3) // ok = false, value =
}
```

### Contributing

Contributions are welcome! 
If you find any bugs or have suggestions for improvements, please open an issue or submit a pull request on GitHub.