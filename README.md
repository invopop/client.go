# Invopop Go Client

The official Invopop API Go Client Library.

## Usage

Start using the Invopop Client in your project by importing the library and initializing the client:

```go

import (
    "context"
    "os"

    "github.com/invopop/client.go/invopop"
)

func main() {
    ctx := context.Background()
    host, _ := os.Getenv("INVOPOP_HOST")
    token, _ := os.Getenv("INVOPOP_TOKEN")
    ic := invopop.New(host, token)

    p := new(invopop.Ping)
    if err := ic.Ping.Fetch(ctx, p); err != nil {
        panic(err.Error())
    }
    fmt.Printf("%v\n", p)
}
```

The Invopop API is split into individual namespaces, these are:

 * `Sequence` - used for generating sequential numbers or codes called `Series`.
 * `Transform` - used to configure `Task`s and `Workflow`s that will be requested to be used when processing `Job`s.
 * `Silo` - for storing GOBL envelopes that will later processed by Jobs and whose task results, if any, will also be stored in the silo service.
