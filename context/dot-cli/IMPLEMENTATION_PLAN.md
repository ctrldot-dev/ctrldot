# Dot CLI Implementation Plan v0.1

## Overview

Build a command-line interface (`dot`) that provides a git-like interface to the Futurematic Kernel. The CLI communicates with the kernel over HTTP and provides commands for reading and writing to the kernel.

## Architecture

### Project Structure
```
cmd/dot/              # Main CLI application
  main.go             # Entry point, command routing
  commands/           # Command implementations
    status.go
    config.go
    show.go
    history.go
    diff.go
    new.go
    role.go
    link.go
    move.go
    ls.go
  client/             # HTTP client for kernel API
    client.go         # HTTP client wrapper
    types.go          # Request/response types
  config/             # Configuration management
    config.go         # Config file handling
    defaults.go       # Default values
  output/             # Output formatting
    text.go           # Text output formatter
    json.go           # JSON output formatter
internal/             # Shared internal packages (if needed)
```

## Implementation Phases

### Phase 1: Foundation & Infrastructure

#### 1.1 Project Setup
- [ ] Create `cmd/dot/` directory structure
- [ ] Set up CLI framework (use `cobra` or `urfave/cli`)
- [ ] Add dependencies to `go.mod`
- [ ] Create main entry point

#### 1.2 Configuration Management
- [ ] Implement config file handling (`~/.dot/config.json`)
- [ ] Support for config keys: `server`, `actor_id`, `namespace_id`, `capabilities`
- [ ] Environment variable overrides:
  - `DOT_SERVER`
  - `DOT_ACTOR`
  - `DOT_NAMESPACE`
  - `DOT_CAPABILITIES`
- [ ] Config get/set commands
- [ ] Config validation

#### 1.3 HTTP Client
- [ ] Create HTTP client wrapper for kernel API
- [ ] Implement all endpoint methods:
  - `GET /v1/healthz`
  - `POST /v1/plan`
  - `POST /v1/apply`
  - `GET /v1/expand`
  - `GET /v1/history`
  - `GET /v1/diff`
- [ ] Error handling and mapping to exit codes
- [ ] Request/response type definitions
- [ ] Timeout handling

#### 1.4 Output Formatting
- [ ] Text formatter (human-readable)
- [ ] JSON formatter (machine-readable)
- [ ] Plan printing (with denies/warns/infos)
- [ ] Operation printing
- [ ] Expand result printing
- [ ] History printing
- [ ] Diff printing

### Phase 2: Connectivity & Config Commands

#### 2.1 Status Command
- [ ] `dot status` implementation
- [ ] Call `GET /v1/healthz`
- [ ] Display server URL, actor ID, namespace, status
- [ ] Support `--json` flag

#### 2.2 Config Commands
- [ ] `dot config get <key>` - Get config value
- [ ] `dot config set <key> <value>` - Set config value
- [ ] Validation for config keys
- [ ] Support for all config keys

#### 2.3 Use Command
- [ ] `dot use <namespace>` - Set namespace in config
- [ ] Update config file
- [ ] Display confirmation

#### 2.4 Whereami Command
- [ ] `dot whereami` - Show resolved config
- [ ] Display server, actor, namespace, capabilities
- [ ] Show environment overrides if any
- [ ] Support `--json` flag

### Phase 3: Read Commands

#### 3.1 Show Command
- [ ] `dot show <node-id>` implementation
- [ ] Support flags: `--depth`, `--asof-seq`, `--asof-time`, `--ns`
- [ ] Call `GET /v1/expand`
- [ ] Format output:
  - Node title + ID
  - Roles (if namespace provided)
  - Links (grouped by type)
  - Materials summary
- [ ] Support `--json` flag

#### 3.2 History Command
- [ ] `dot history <target>` implementation
- [ ] Support `--limit` flag
- [ ] Call `GET /v1/history`
- [ ] Format output: seq, occurred_at, actor_id, op_id, class, change summary
- [ ] Support `--json` flag

#### 3.3 Diff Command
- [ ] `dot diff <a> <b> <target>` implementation
- [ ] Support seq integers and `now` alias
- [ ] Resolve `now` via `/v1/history?target=<target>&limit=1`
- [ ] Call `GET /v1/diff`
- [ ] Format output: created/retired changes
- [ ] Support `--json` flag

#### 3.4 Ls Command (Optional)
- [ ] `dot ls <node-id>` implementation
- [ ] Implement via expand depth=1
- [ ] Filter children where `link.type==PARENT_OF` and `link.from==<node-id>`
- [ ] List child nodes
- [ ] Support `--asof-seq`, `--asof-time`, `--ns` flags
- [ ] Support `--json` flag

### Phase 4: Write Commands

#### 4.1 Plan/Apply Workflow
- [ ] Implement shared plan/apply logic
- [ ] Call `POST /v1/plan`
- [ ] Print plan + policy report
- [ ] Check for denies → exit code 2
- [ ] Support `--dry-run` flag → exit after plan
- [ ] Prompt for confirmation (unless `--yes`)
- [ ] Call `POST /v1/apply`
- [ ] Print operation result
- [ ] Handle errors (conflict, policy denied, etc.)

#### 4.2 New Node Command
- [ ] `dot new node "<title>"` implementation
- [ ] Support `--meta k=v ...` flags
- [ ] Support `--ns` flag
- [ ] Create `CreateNode` intent
- [ ] Use plan/apply workflow
- [ ] Support `--dry-run` and `--yes` flags
- [ ] Support `--json` flag

#### 4.3 Role Assign Command
- [ ] `dot role assign <node-id> <role>` implementation
- [ ] Support `--ns` flag
- [ ] Create `AssignRole` intent
- [ ] Use plan/apply workflow
- [ ] Support `--dry-run` and `--yes` flags
- [ ] Support `--json` flag

#### 4.4 Link Command
- [ ] `dot link <from> <type> <to>` implementation
- [ ] Support `--ns` flag
- [ ] Create `CreateLink` intent
- [ ] Use plan/apply workflow
- [ ] Support `--dry-run` and `--yes` flags
- [ ] Support `--json` flag

#### 4.5 Move Command
- [ ] `dot move <child> --to <parent>` implementation
- [ ] Support `--ns` flag
- [ ] Check if kernel supports `Move` intent
- [ ] If supported: use `Move` intent
- [ ] If not: fetch current parent, retire link, create new link
- [ ] Use plan/apply workflow
- [ ] Support `--dry-run` and `--yes` flags
- [ ] Support `--json` flag

### Phase 5: Global Flags & Polish

#### 5.1 Global Flags
- [ ] `--server <url>` - Override server URL
- [ ] `--actor <id>` - Override actor ID
- [ ] `--ns <namespace>` - Override namespace
- [ ] `--cap <capabilities>` - Override capabilities (comma-separated)
- [ ] `--json` - JSON output mode
- [ ] `-n, --dry-run` - Plan only (mutations)
- [ ] `-y, --yes` - Skip confirmation (mutations)

#### 5.2 Error Handling
- [ ] Map kernel error codes to exit codes:
  - `POLICY_DENIED` → exit code 2
  - `CONFLICT` → exit code 3
  - `VALIDATION` → exit code 1
  - `NOT_FOUND` → exit code 1
  - `INTERNAL`/others → exit code 4
- [ ] Client-side validation errors → exit code 1
- [ ] Network errors → exit code 4
- [ ] Clear error messages

#### 5.3 Output Polish
- [ ] Consistent formatting across commands
- [ ] Stable output (suitable for scripts)
- [ ] Important data first (IDs, seqs, denies)
- [ ] No ANSI colors by default
- [ ] Proper JSON formatting

#### 5.4 Testing
- [ ] Unit tests for config management
- [ ] Unit tests for HTTP client
- [ ] Unit tests for output formatters
- [ ] Integration tests with kernel (requires kernel running)
- [ ] Test all commands
- [ ] Test error handling
- [ ] Test exit codes

### Phase 6: Documentation & Build

#### 6.1 Documentation
- [ ] README for dot-cli
- [ ] Command reference
- [ ] Examples
- [ ] Installation instructions

#### 6.2 Build & Distribution
- [ ] Makefile target for building
- [ ] Cross-platform builds
- [ ] Binary distribution
- [ ] Installation script (optional)

## Technical Decisions

### CLI Framework Choice
**Recommendation: Use `cobra`**
- Mature and widely used
- Good support for subcommands
- Built-in help generation
- Flag handling
- Easy to extend

### Configuration Storage
- Store at `~/.dot/config.json`
- Use `encoding/json` for parsing
- Create directory if it doesn't exist
- Validate JSON structure

### HTTP Client
- Use standard `net/http` package
- Implement timeout handling
- Proper error wrapping
- Response parsing

### Output Formatting
- Separate formatters for text and JSON
- Use interfaces for extensibility
- Consistent structure across commands

## Dependencies

```go
require (
    github.com/spf13/cobra v1.8.0  // CLI framework
    // Standard library:
    // - encoding/json
    // - net/http
    // - os
    // - path/filepath
    // - fmt
    // - strings
    // - time
)
```

## File Structure Details

```
cmd/dot/
├── main.go                    # Entry point, command registration
├── commands/
│   ├── status.go              # dot status
│   ├── config.go              # dot config get/set
│   ├── use.go                  # dot use
│   ├── whereami.go             # dot whereami
│   ├── show.go                 # dot show
│   ├── history.go              # dot history
│   ├── diff.go                 # dot diff
│   ├── ls.go                   # dot ls (optional)
│   ├── new.go                  # dot new node
│   ├── role.go                 # dot role assign
│   ├── link.go                 # dot link
│   └── move.go                 # dot move
├── client/
│   ├── client.go               # HTTP client wrapper
│   └── types.go                 # Request/response types matching kernel API
├── config/
│   ├── config.go               # Config file management
│   └── defaults.go             # Default values
└── output/
    ├── formatter.go            # Formatter interface
    ├── text.go                 # Text formatter
    └── json.go                 # JSON formatter
```

## Implementation Order

1. **Phase 1** - Foundation (config, client, output)
2. **Phase 2** - Connectivity commands (status, config, use, whereami)
3. **Phase 3** - Read commands (show, history, diff, ls)
4. **Phase 4** - Write commands (new, role, link, move)
5. **Phase 5** - Polish (global flags, error handling, testing)
6. **Phase 6** - Documentation and build

## Acceptance Criteria

- [ ] All commands from spec implemented
- [ ] All global flags work
- [ ] Config file management works
- [ ] Environment variable overrides work
- [ ] Plan/apply workflow works for all mutations
- [ ] Policy denies exit with code 2
- [ ] Dry-run works (plan only)
- [ ] JSON output works for all commands
- [ ] Text output is stable and readable
- [ ] Exit codes match specification
- [ ] Error handling is comprehensive
- [ ] Commands work with running kernel

## Testing Strategy

1. **Unit Tests**
   - Config management
   - HTTP client (with mock server)
   - Output formatters
   - Command parsing

2. **Integration Tests**
   - Test against running kernel
   - Test all commands end-to-end
   - Test error scenarios
   - Test exit codes

3. **Manual Testing**
   - Test user experience
   - Test output formatting
   - Test edge cases

## Next Steps After v0.1

- Interactive TUI mode
- AI chat mode
- Batch operations
- More output formats (YAML, table)
- Command aliases
- Shell completion
- Progress indicators for long operations
