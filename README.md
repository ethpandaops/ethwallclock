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

### Variable slot durations (EIP-8198)

Chains using a slot duration schedule (EIP-8198 `SLOT_DURATION_SCHEDULE`) can
pass the schedule instead of a single duration. Slot/epoch numbers, time
windows and the `OnSlotChanged`/`OnEpochChanged` callbacks follow it.

```go
wallclock, err := ethwallclock.NewEthereumBeaconChainWithSchedule(genesisTime, []ethwallclock.SlotDuration{
  {Epoch: 0, Duration: 12 * time.Second},
  {Epoch: 100000, Duration: 11 * time.Second},
}, 32)
```
