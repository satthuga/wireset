---
title: Getting Started
---

# Getting Started with Efficient Wireset

Welcome to Efficient Wireset! This guide will help you get started with our collection of common wireset configurations for developing applications on Shopify.

## Prerequisites

Before you begin, make sure you have the following installed:

- Go
- Google Wire

## Installation

To use this module in your project, you need to add it as a dependency. Run the following command in your project directory:

```bash
go get github.com/aiocean/wireset
````

This will download the module and add it to your go.mod file.

## Using Wireset in Your Code

After adding the wireset module to your project, you can use it in your Go files. First, import the wireset module:

```go
import "github.com/aiocean/wireset"
```

Then, you can use the wire.Build function to generate the necessary code for dependency injection. For example, if you want to use the NewDgraphSvc function from the wireset module, you can do it like this:

```go
func InitializeDgraphSvc() (*dgo.Dgraph, error) {
    wire.Build(wireset.NewDgraphSvc)
    return &dgo.Dgraph{}, nil
}
```

In this example, InitializeDgraphSvc is a placeholder function that you need to replace with your own initialization function. When you run the wire command, Wire will generate the necessary code to initialize a dgo.Dgraph instance using the NewDgraphSvc function from the wireset module.  

## Generating Code with Wire

After you've set up your code to use Wire, you can generate the necessary code by running the wire command in your project directory:

```bash
wire
```

This will generate a wire_gen.go file with the necessary code for dependency injection. 

You now can use the InitializeDgraphSvc function to initialize a dgo.Dgraph instance:

```go
func main() {
    dgraph, err := InitializeDgraphSvc()
    if err != nil {
        log.Fatal(err)
    }
    // Use dgraph
}
```
