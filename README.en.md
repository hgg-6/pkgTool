# pkg_tool

An open-source toolkit project based on Gitee, providing encapsulation of logging functionalities to help developers quickly integrate logging capabilities into their Go projects.

## Project Structure

- `go.mod` and `go.sum`: Go module configuration files.
- `log/zeroLog/`: Encapsulates logging interfaces and implementations based on [zerolog](https://github.com/rs/zerolog).

## Features

- Provides a unified logging interface `Zlogger`.
- Supports multiple logging levels: Info, Error, Debug, Warn.
- Offers log context support for easier tracking of log information.

## Component Description

### `log/zeroLog/logtest.go`

- `Zlogger`: Defines the logging interface.
- `Zlog`: Concrete structure implementing the `Zlogger` interface.
- `NewZlog`: Creates a new instance of `Zlog`.
- `Info/Error/Debug/Warn`: Outputs logs at different levels.
- `With`: Adds contextual information to logs.

### `log/zeroLog/logtest_test.go`

- `TestInitLog`: Unit test for logging initialization functionality.

## Usage

1. Install dependencies:

   ```bash
   go get github.com/rs/zerolog
   ```

2. Initialize the logger:

   ```go
   logger := NewZlog(&zerolog.Logger{})
   ```

3. Use the logger:

   ```go
   logger.Info().Msg("This is an info log")
   logger.Error().Msg("This is an error log")
   ```

## Contribution Guide

Code contributions and documentation improvements are welcome. Please follow these steps:

1. Fork this repository.
2. Create a new branch.
3. Commit your changes.
4. Submit a Pull Request.

## License

This project is licensed under the MIT License. For details, please refer to the LICENSE file in the project root directory.

---

The above is a basic README file generated based on the existing code structure. You can expand upon it with more detailed information according to the actual functionality.