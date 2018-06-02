# fio [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/jbub/fio) [![Build Status](https://travis-ci.org/jbub/fio.svg?branch=master)](https://travis-ci.org/jbub/fio) [![Coverage Status](https://coveralls.io/repos/github/jbub/fio/badge.svg?branch=master)](https://coveralls.io/github/jbub/fio?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/jbub/fio)](https://goreportcard.com/report/github.com/jbub/fio)

API client for Fio Banka written in Go.

## Install

```bash
go get github.com/jbub/fio
```

## Docs

https://godoc.org/github.com/jbub/fio

## Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    "context"

    "github.com/jbub/fio"
)

func main() {
    client := fio.NewClient("mytoken", nil)

    opts := fio.ByPeriodOptions{
        DateFrom: time.Now(),
        DateTo:   time.Now(),
    }

    resp, err := client.Transactions.ByPeriod(context.Background(), opts)
    if err != nil {
        log.Fatal(err)
    }

    for _, tx := range resp.Transactions {
        fmt.Println(tx.ID)
    }
}
```
