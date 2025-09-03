## Intro

There are billions of libraries and tools available today, but AI systems can’t easily leverage them in structured workflows. AtomOS aims to change that: any library that wants to be “agent-friendly” can include a standard agentic_support.yaml file, provide a CLI interface, and release a compiled binary. Once these three pieces are added, the library becomes fully accessible to agentic systems through a unified runtime. There will undoubtedly be technical challenges, but this approach has the potential to enable far more powerful and deterministic AI workflows.

## Main Components

### Blocks

What constitutes a “block”? A block is any piece of software that meets all four of the following criteria:

1. **`agentic_support.yaml` file** – defines the block’s inputs, outputs, and CLI entries.
2. **CLI interface** – exposes the library’s functionality in a standardized way that agentic systems can call.
3. **Compiled binary** – the executable version of the software for the target platform.
4. **GitHub-hosted** – the release binary must be attached to a GitHub release, and the `agentic_support.yaml` file must be located at the root of the repository.

Blocks meeting these requirements can be imported and composed into workflows by agentic systems using AtomOS.

### Workflows

The `agentic_support.yaml` file defines LSP-style entries, which describe the CLI commands exposed by the block and how to use them. Since every block must implement a CLI, these entries provide the standardized interface for communication.

Using this information, AtomOS can compose blocks into workflows: outputs from one block can be passed as inputs to another, forming a graph-like structure. This enables the creation of complex, deterministic workflows that agentic systems can execute reliably.

### Binary Management

We should rely on existing solutions wherever possible. Handling versioning, checksums, and reliable downloads is complex and error-prone, so implementing it from scratch is unnecessary. Instead, AtomOS can leverage libraries like go-getter for fetching binaries from GitHub releases, and go-update for safely managing updates. Combined with a local cache per OS/ARCH/version, this approach ensures reproducible, verifiable, and maintainable block binaries with minimal custom code.

#### Concerners

##### **1. Executing many binaries is heavier than in-process calls**

- Each CLI call spawns a **new process**, which involves:

  - OS process creation overhead
  - Memory allocation for the process
  - Loading the binary (paged in as needed)

- Compared to calling functions in the same process, this is slower and slightly more resource-intensive.

##### **2. How heavy it actually is**

- For **small CLI utilities**, the overhead is usually negligible — a few milliseconds per call.
- For **blocks that do heavy computation or I/O**, the cost of spawning a process is minor relative to the work done.
- For **very high-frequency calls** or **tight loops**, the overhead can add up.

## Vision

1. **Universal Package Manager**
   AtomOS aims to provide a unified system for discovering, fetching, and versioning blocks from GitHub. By standardizing metadata through `agentic_support.yaml`, it ensures that any compliant library or tool can be integrated seamlessly, with automatic handling of dependencies, versions, and binaries.

2. **Universal Runtime Executor**
   AtomOS will offer a runtime capable of executing any block, regardless of programming language or platform, in a deterministic and reproducible way. This includes managing binaries, invoking CLI commands, and orchestrating multi-step workflows efficiently.

3. **Composable Workflows Across Any Software**
   The ultimate goal is to enable the creation of complex workflows by combining any software that follows AtomOS standards. Outputs from one block can flow into another, forming graph-like workflows that agentic systems can reliably execute, unlocking the full potential of existing OSS libraries and tools.
