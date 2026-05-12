# Code Review Action Items

## 1. Naming Fixes

- [x] `drivers/zh07/zh07q.go:104` — Fix `especifics` typo → `specifics`
- [x] `drivers/sps30/sps30.go:112` — Fix `especifics` typo → `specifics`
- [x] `drivers/zh07/helpers.go:9` — Rename `tempq` → `checksum`
- [x] `drivers/zh07/helpers.go:17` — Rename `byteToInt` → `bytesToUint16BE`
- [x] `drivers/sps30/helpers.go:28` — Rename `byteArrayToFloat32` → `bytesToFloat32`
- [x] `drivers/sps30/helpers.go:24` — Rename `crc` → `crc8Checksum`

## 2. Missing Comments / Godoc

- [x] `domain/transport_provider.go` — Add godoc for `TransportProvider` interface
- [x] `domain/sensor_provider.go` — Add godoc for `SensorProvider` interface
- [x] `drivers/sps30/sps30.go:36` — Document why `Init()` body is empty
- [x] `drivers/sps30/sps30.go:113` — Document `ReadArticleCode` (non-obvious term)
- [x] `internal/helpers/helpers.go:17` — Add godoc for `WithRetry` function
- [x] `internal/helpers/helpers.go:12` — Clarify `CountAs` comment ("nil = everything counts" → "nil means all errors count toward retry limit")
- [x] `internal/helpers/helpers.go:13` — Fix "father ctx" → "parent context"
- [x] `drivers/zh07/zh07i.go:30` — Document default transport as test no-op
- [x] `drivers/zh07/zh07q.go:30` — Document default transport as test no-op

## 3. Dead Code Removal

- [x] `drivers/zh07/helpers.go:22-30` — Remove unused exported `toHex` function
- [x] `domain/reading.go:12,17` — Remove `MassPM4` and `NumberPM4` (never populated) or document as sensor-specific

## 6. New Findings

### 6.1 Bugs

- [x] `internal/helpers/helpers.go:21-27` — `context.WithCancel(ctx)` at line 21 leaked when timeout context is created at line 25 — first cancel never called

### 6.2 Code Quality

- [x] `drivers/zh07/helpers.go:4` — `calculateChecksum` takes `*[]byte` (pointer to slice) — unnecessary indirection, should take `[]byte`
- [x] `drivers/zh07/zh07i.go:69,82,87` — `Read()` returns `&ReadingEvent{}` with nil Reading on bad frames — caller can't distinguish "no data yet" from "valid empty frame" (resolved: return `ErrInvalidFrame` + nil-guard in `Run()`)
- [x] `drivers/zh07/zh07.go:30` — `New()` returns `nil` for unknown mode — silent failure, no error returned
- [x] `drivers/sps30/sps30.go:207-211` — `IsDataReady()` returns `int` (-1/0/1) — ambiguous return, should return `(bool, error)`
- [x] `internal/transport/i2c/i2c.go:37` — Format string `"i2c write: %X\n%w"` — literal `\n` prints as text, not newline
- [x] `drivers/sps30/sps30.go:137-152` — `ReadSerial()` doesn't handle NUL-terminated strings like `ReadArticleCode()` — could include trailing zeros

### 6.3 Incomplete Features

- [ ] `domain/reading.go` — `Timestamp` and `SensorID` fields never populated by any driver
  - Plan: add optional `id string` param to each `New()` constructor, default to driver type name ("sps30", "zh07-q", "zh07-i") when empty
  - Plan: set `Timestamp: time.Now()` and `SensorID: s.id` in each driver's `Read()` / `ReadMeasurement()` at the point of creating the reading

### 6.4 Typos

- [ ] `drivers/zh07/zh07_test.go:31,37,42` — "sucess" → "success"
