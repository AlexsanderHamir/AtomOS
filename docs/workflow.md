## Intro

The workflow compiles and runs the specified config on the {name}\_workflow_atoms.yaml file.

### Input

1. For nodes with no incoming edges, the input comes from a file.
2. For nodes with incoming edges its output can stay on a named buffer if not too big for in memory.

### Thoughts

I think that `to_block` and `to_entry` are not necessary, and that the connections should be made based on who uses the input of who instead.

### Connection model

Connections are inferred by matching outputs to inputs:

- A block declares what it produces via `output` and what it consumes via `input`.
- An edge is created from the producer's `from_block` to every consumer whose `input` equals that `output`.
- There is no need to specify `to_block` or `to_entry`.

Root/source connections are the ones without any `input` set. These represent the initial ingestion of data from files or other external sources and are executed by piping the `source` file into the block's binary.

### YAML schema (relevant fields)

- `blocks[]`: list of blocks with `name`, `version`, `github`, `force`.
- `connections[]` items:
  - `from_block`: producer block name
  - `from_entry`: entry within the producer that emits the output
  - `output`: logical name for the produced data
  - `input` (optional): logical name this block consumes; if omitted, this is a root/source
  - `source` (optional): path used for root/source connections

### Example

```yaml
connections:
  - from_block: filemanager
    from_entry: list
    output: file_list
    source: path/to/file

  - from_block: textprocessor
    from_entry: count
    output: statistics
    input: file_list

  - from_block: sysmonitor
    from_entry: system
    output: system_info
    input: statistics
```

In this example, an edge is created from `filemanager -> textprocessor` because `textprocessor.input == filemanager.output (file_list)`, and another from `textprocessor -> sysmonitor` because `sysmonitor.input == textprocessor.output (statistics)`.
