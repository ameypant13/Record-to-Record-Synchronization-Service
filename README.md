# Record-to-Record-Synchronization-Service


## Schema Transformer & Validator

A robust Golang engine for bi-directional, record-by-record **schema transformation and validation**—ideal for data synchronization between internal apps and external platforms/CRMs.

---

### ✨ Features

- **JSON Schema (draft-07) validation** with [qri-io/jsonschema](https://github.com/qri-io/jsonschema)
- **Configurable field mapping**: supports field renames, type coercion, enum ↔ bool conversions
- **Bidirectional**: easily map in both directions (internal ↔ external)
- **Detailed error reporting**: robust validation and transform errors
- **Easily extensible**: just add schemas and config for new objects or providers

---

## 🚀 Quick Start

### 1. Prerequisites

- [Go 1.18+](https://golang.org/dl/)
- [qri-io/jsonschema](https://github.com/qri-io/jsonschema)

```sh
go get github.com/qri-io/jsonschema

## Run the service
```sh
go run main.go
```