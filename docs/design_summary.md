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

#### Concerners

##### **1. Executing many binaries is heavier than in-process calls**

- Each CLI call spawns a **new process**, which involves:

  - OS process creation overhead
  - Memory allocation for the process
  - Loading the binary (paged in as needed)

- Compared to calling functions in the same process, this is slower and slightly more resource-intensive.


### Vision

The ability to connect any function from one piece of software to any function in another unlocks powerful customization. While it may introduce some performance overhead, that’s a technical challenge we can mitigate.