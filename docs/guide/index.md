# Get started

I assume you have a working knowledge of Wireset - a dependency injection framework for Golang by Google.

We will start from scratch and build a simple application that will demonstrate the use of Wireset.

## Prerequisites

- Go 1.2 or higher
- [Wireset](https://github.com/google/wireset)

## Implement

Create your main endpoint in `cmd/app/main.go`:

```
├── cmd
|   └── app
|       └── main.go
|       └── wire.go
|── go.mod
```

In wire.go file, we will use `wire.Build` to create an initializer function for our application.


```go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/aiocean/aibridge-server/feature"
	"github.com/aiocean/wireset"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/server"
	"github.com/google/wire"
)

func InitializeServer() (*server.Server, func(), error) {
	wire.Build(
		// core
		wireset.NormalAppWireset,
		// config
		configsvc.EnvWireset,
		pubsub.GoroutineWireset,
	)
	return &server.Server{}, nil, nil
}
```

In this function, we declare all the providers and injectors that we need to create server instance.

Wireset provides two type of wireset to create normal and shopify application.

- Normal application: `wireset.NormalAppWireset`
- Shopify application: `wireset.ShopifyAppWireset`

In this example, we create a normal application. This wireset need some providers to works, such as: config, and pubsub.

To generate the code that golang can use to buld the binary, we need to run the following command:

```bash
wire ./cmd/app
````

in file main.go, we will use the initializer function to create server instance and start it.

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	server, cleanup, err := InitializeServer()
	if err != nil {
		log.Panic("failed to init server", err)
	}

	defer func() {
		log.Println("starting cleanup")
		cleanup()
		log.Println("finished cleanup")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()
	doneChan := server.Start(ctx)

	for {
		select {
		case reason := <-doneChan:
			if reason != nil {
				log.Println("context error:", reason)
				return
			}

			log.Println("server done without reason")
		case <-quit:
			log.Println("received quit signal")
			return
		}
	}
}
```

