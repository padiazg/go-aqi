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

## 4. Bug Fix
- [ ] `drivers/sps30/sps30.go:193` — Fix `in[27]` used twice → likely `in[28]` for `NumberPM05`

## 5. Minor Code Quality
- [ ] `internal/helpers/helpers.go:18-24` — Fix leaked `context.WithCancel(ctx)` when timeout > 0
- [ ] `drivers/zh07/zh07i.go:52` — Document what "i" stands for in `ZH07i`
- [ ] `drivers/zh07/zh07q.go:51` — Document what "q" stands for in `ZH07q`
