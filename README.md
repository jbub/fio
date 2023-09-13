# fio 
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/jbub/fio) 
[![Build Status](https://github.com/jbub/fio/actions/workflows/go.yml/badge.svg)](https://github.com/jbub/fio/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbub/fio)](https://goreportcard.com/report/github.com/jbub/fio)

API client for Fio Banka written in Go.

## Install

```bash
go get github.com/jbub/fio
```

## Docs

https://pkg.go.dev/github.com/jbub/fio

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
