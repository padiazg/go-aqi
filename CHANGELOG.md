# Changelog

## [Unreleased]

### Added
- Initial project structure with domain, drivers (ZH07B, SPS30), and internal transport layer
- ZH07B driver with Initiative Upload (`ZH07i`) and Question/Answer (`ZH07q`) modes
- SPS30 driver with I2C transport, measurement start/stop, data-ready polling, and sensor identification
- I2C and Serial transport abstractions
- Retry helper with configurable timeout, interval, and count-as filtering
- CRC-8 checksum support for SPS30 sensor protocol
