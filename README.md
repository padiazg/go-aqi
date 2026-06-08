# go-aqi

[![Go Reference](https://pkg.go.dev/badge/github.com/padiazg/go-aqi.svg)](https://pkg.go.dev/github.com/padiazg/go-aqi)
[![Go Report Card](https://goreportcard.com/badge/github.com/padiazg/go-aqi)](https://goreportcard.com/report/github.com/padiazg/go-aqi)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Status:** Pre-release / Work in progress

A Go library for reading air quality data from particulate matter sensors. Provides a unified interface for querying PM concentrations (PM1, PM2.5, PM4, PM10) from different sensor models.

## Supported Sensors

| Sensor | Interface | Mass Concentration | Number Concentration |
|--------|-----------|-------------------|---------------------|
| [ZH07B](https://github.com/padiazg/go-aqi/tree/main/drivers/zh07) | Serial (UART) | PM1, PM2.5, PM10 | PM1, PM2.5, PM10 |
| [SPS30](https://github.com/padiazg/go-aqi/tree/main/drivers/sps30) | I2C | PM1, PM2.5, PM4, PM10 | PM0.5, PM1, PM2.5, PM4, PM10 |

## ZH07B Modes

The ZH07B driver supports two communication modes:

- **Initiative Upload Mode (`ZH07i`)** - The sensor automatically pushes data; the driver reads frames as they arrive.
- **Question/Answer Mode (`ZH07q`)** - The driver queries the sensor and receives a response with checksum validation.

## SPS30 Features

- Measurement start/stop control
- Data-ready polling with retry logic
- Sensor identification (article code, serial number)
- Auto-cleaning interval configuration
- Reset command support

## Installation

```bash
go get github.com/padiazg/go-aqi
```

## Quick Start

### ZH07B (Question/Answer Mode)

```go
import (
    "github.com/padiazg/go-aqi/drivers/zh07"
)

config := &zh07.Config{
    Transport: serialTransport, // your serial transport
    Interval:  1 * time.Second,
    Mode:      zh07.ModeQA,
}

sensor, err := zh07.New(config)
if err != nil {
    log.Fatal(err)
}
sensor.Init(ctx)

for event := range sensor.Run(ctx) {
    if event.Err != nil {
        log.Fatal(event.Err)
    }
    fmt.Printf("PM2.5: %.2f µg/m³\n", event.Reading.NumberPM25)
}
```

### SPS30

```go
import (
    "github.com/padiazg/go-aqi/drivers/sps30"
)

i2cBus, _ := i2c.OpenI2C(1, 0x69) // I2C bus 1, address 0x69

sensor := sps30.New(i2cBus, 1*time.Second)
sensor.Init(ctx)

for event := range sensor.Run(ctx) {
    if event.Err != nil {
        log.Fatal(event.Err)
    }
    fmt.Printf("PM2.5: %.2f µg/m³\n", event.Reading.MassPM25)
}
```

## Project Structure

```
.
├── domain/          # Core interfaces and types
│   ├── sensor_provider.go
│   ├── transport_provider.go
│   └── reading.go
├── drivers/         # Sensor driver implementations
│   ├── zh07/        # ZH07B particulate matter sensor
│   └── sps30/       # SPS30 laser scattering sensor
├── transport/       # Communication transport abstractions
│   ├── serial/      # UART serial transport
│   └── i2c/         # I2C bus transport
└── pkg/             # Shared utilities
    └── helpers/     # Retry logic and common helpers
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
.