# EthWallclock

[![Go Report Card](https://goreportcard.com/badge/github.com/ethpandaops/ethwallclock)](https://goreportcard.com/report/github.com/ethpandaops/ethwallclock)
[![Release](https://img.shields.io/github/v/release/ethpandaops/ethwallclock)](https://github.com/ethpandaops/ethwallclock/releases)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

An Ethereum Wallclock Go package

## Installation

```bash
go get github.com/ethpandaops/ethwallclock
```

## Usage

```go
package main

import (
  "fmt"
  "time"

  "github.com/ethpandaops/ethwallclock"
)
 
func main() {
  genesisTime, _ := time.Parse(time.RFC3339, "2020-01-01T12:00:23Z")

  wallclock := ethwallclock.NewEthereumBeaconChain(genesisTime, 12*time.Second, 32)

  slot, epoch, err := wallclock.Now()
  if err != nil {
    panic(err)
  }

  fmt.Println("Slot: ", slot)
  fmt.Println("Epoch: ", epoch)
}
```
