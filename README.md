
# Invopop Go Client

The official Invopop API Go Client Library.

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/badge-maker?url=https%3A%2F%2Fdeepwiki.com%2Finvopop%2Fclient.go)

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
    token, _ := os.Getenv("INVOPOP_TOKEN")
    ic := invopop.New()
    ic.SetAuthToken(token)

    p := new(invopop.Ping)
    if err := ic.Utils().Ping(ctx, p); err != nil {
        panic(err.Error())
    }
    fmt.Printf("%v\n", p)
}
```

The Invopop API is split into individual namespaces, these are:

- `Utils` - for ensuring you can connect correctly with the Invopop servers and your credentials are correct.
- `Sequence` - used for generating sequential numbers or codes called `Series`.
- `Transform` - used to configure `Integration`s and `Workflow`s that will be requested to be used when processing `Job`s.
- `Silo` - for storing GOBL envelopes ready to send to integrations via jobs whose output may be stored as attachments.
- `Access` - mostly used by apps to get access tokens and manage enrollment data.
