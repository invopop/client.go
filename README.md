# client
API Client Library

## Test

### Sequence client

Run the following command to run the sequence tests

```shell
go test -v api/sequence/sequence_test.go
```

If the API server has issues with its SSL certificate, use the following
configuration to use a insecure conection (don't push, just for dev and test).

```
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    },
```
