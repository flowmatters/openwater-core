# OpenWater Core Copilot Instructions

## Project Overview
OpenWater Core is the computational engine for hydrological modeling written in Go. It provides high-performance simulation capabilities with C bindings for integration with Python and other languages.

## Technology Stack
- **Language**: Go 1.12+
- **Build**: Go modules (go.mod)
- **Key Dependencies**:
  - HDF5 for data I/O
  - Protocol Buffers for data serialization
  - YAML for configuration

## Project Structure
- `cmd/`: Command-line tools (ow-sim, ow-inspect, ows-ensemble, ow-single)
- `sim/`: Core simulation engine
- `models/`: Hydrological model implementations
- `io/`: Input/output handling (HDF5, etc.)
- `graph/`: Model graph/network structures
- `util/`: Utility functions
- `test/`: Test files
- `libopenwater.so`: C-compatible shared library for Python integration

## Model implementations

Models are implemented as Go structs that satisfy the `Model` interface. They are composable and can be connected in a graph structure to represent complex hydrological systems. Each model has its own parameters and state, and they exchange data through defined input/output ports.

Models are typically implemented as a single function and an accompanying OW-SPEC comment that is processed by the build system (ow-specgen) to generate necessary boilerplate code for integration with the simulation engine.

## Code Style
- Follow standard Go conventions (gofmt, golint)
- Use descriptive variable names
- Prefer composition over inheritance
- Keep functions focused and testable
- Document exported functions and types

## Testing
- Write table-driven tests where appropriate
- Use testify for assertions
- Test both success and error paths
- Keep tests close to the code they test

## Building
- Standard Go build: `go build`
- Use go modules for dependency management
- CGO is enabled for HDF5 bindings
- Produces shared library for Python integration

## Key Patterns
- Models are composable units in a graph
- Data flows through nodes in the graph
- HDF5 used for efficient time series I/O
- Support for ensemble simulations

## Integration
- Exposes C API via libopenwater.so
- Python bindings consume the shared library and process models files and results (h5)
- Keep C API stable for backwards compatibility
- Document any breaking changes to the API
