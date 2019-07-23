# Solgen

## Getting Started
### Prerequisites
* Go == 1.12

### install
```
go get github.com/frostornge/solgen
```

### build
```
go build
```

## Usage
```
Solgen is a tool for generate solidity binds.
This application helps to generate go/proto bind of solidity.

Usage:
  solgen [flags]

Flags:
  -h, --help            help for solgen
  -i, --input string    Input path (default "http://localhost:8500")
  -o, --output string   Output path (default "./out/")
  -t, --type string     Bind type (default "go")
      --version         version for solgen
```