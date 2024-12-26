# Simple Cache

Simple interface for caching

## Installation

scache requires Go 1.23 or later.

```
go get github.com/safr/scache
```

## Usage

```go
package main

import (
	"log"
	"time"

	cache "github.com/safr/scache"
)

func main() {
	cache := cache.New(10)
	if cache == nil {
		log.Fatal("New() is nil")
	}

	cache.StartEvictionTicker(1 * time.Minute)

	if err := cache.Set("key1", "value1", 1*time.Hour); err != nil {
		log.Fatal(err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("value: %s \n", value)

	if err := cache.Flush(); err != nil {
		log.Fatal(err)
	}

	if cache.Contains("key1") {
		log.Fatal(err)
	}
}

```

### Makefile

```sh
// Setup
$ make setup

// Format all go files
$ make format

//Run linters
$ make lint

// Run tests
$ make test
```

## License

This project is released under the MIT licence. See [LICENSE](https://github.com/safr/scache/blob/main/LICENSE) for more details.
