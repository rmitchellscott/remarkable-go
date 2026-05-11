# remarkable-go

[![rm1](https://img.shields.io/badge/rM1-supported-green)](https://remarkable.com/store/remarkable)
[![rm2](https://img.shields.io/badge/rM2-supported-green)](https://remarkable.com/store/remarkable-2)
[![rmpp](https://img.shields.io/badge/rMPP-supported-green)](https://remarkable.com/store/overview/remarkable-paper-pro)
[![rmppmove](https://img.shields.io/badge/rMPPMove-supported-green)](https://remarkable.com/products/remarkable-paper/pro-move)
[![rmppure](https://img.shields.io/badge/rMPPure-supported-green)](https://remarkable.com/products/remarkable-paper/pure)

Go library for managing reMarkable tablet internals.

## Packages

- **device** — identify device type (RM1, RM2, Paper Pro variants) and CPU architecture
- **partition** — inspect and switch A/B root partitions, with encryption-aware safety checks
- **update** — install SWU firmware images via `swupdate-from-image-file`
- **version** — parse and compare reMarkable OS version strings
- **executor** — command execution abstraction (local, SSH, dry run)
- **filesystem** — file I/O abstraction (local, SFTP, mock) with mount support

## Install

```
go get github.com/rmitchellscott/remarkable-go
```

## License
Copyright (C) 2026 Mitchell Scott

LGPL-3.0
