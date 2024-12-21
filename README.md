<div align="center">

```
   ____                   _                    ____ _     ___ 
  / __ \                 | |                  / ___| |   |_ _|
 | |  | |_   _  __ _ _ __| |_ _   _ _ __ ___ | |   | |    | | 
 | |  | | | | |/ _' | '__| __| | | | '_ ' _ \| |   | |    | | 
 | |__| | |_| | (_| | |  | |_| |_| | | | | | | |___| |___ | | 
    \___\_\\__,_|\__,_|_|   \__|\__,_|_| |_| |_|\____|_____|___|  
```

[![Tests](https://github.com/andreivisan/quantum_cli/actions/workflows/tests.yml/badge.svg)](https://github.com/andreivisan/quantum_cli/actions/workflows/tests.yml)

</div>

## Prerequisites

- Go 1.21 or later
- [Ollama](The CLI tool will guide you through the installation if you don't have it)
- Python 3.10 or later

### Get the Python server

Because LangChain only has an official library for Python and JS, we need to run a Python server to implement Chain of Thought and communicate with the AI.

Please follow the instructions [here](https://github.com/andreivisan/quantum_server) to get the server running.

## Installation

### Option 1: Go Install

If you have Go installed, you can install CharmLlama using:

```bash
go install github.com/andreivisan/quantum_cli@latest
```

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Then reload your shell configuration:

```bash
source ~/.zshrc
```

### Option 2: Build from source

- Clone the repository

```bash
git clone https://github.com/andreivisan/quantum_cli.git
cd quantum_cli
```

- Build the binary

```bash
go build -o quantum_cli
```

- Add the binary to your PATH

## Usage

If the binary is in your PATH, you can run it directly:

```bash
quantum_cli
```

If you want to run it from the current directory, you can use:

```bash
./quantum_cli
```
