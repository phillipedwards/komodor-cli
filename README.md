# komodor

A command-line interface for the [Komodor](https://komodor.com) platform API.

## Installation

### Prerequisites

- Go 1.23 or later

### Build from source

```bash
git clone https://github.com/phillipedwards/komodor-cli.git
cd komodor-cli
make build
```

The binary is written to `./bin/komodor`. Move it somewhere on your `$PATH`:

```bash
mv ./bin/komodor /usr/local/bin/komodor
```

## Authentication

The CLI resolves your API key in this order (first match wins):

1. `--api-key` flag
2. `KOMODOR_API_KEY` environment variable
3. Config file at `~/.config/komodor/config.yaml`

**Recommended:** save your key to the config file once and forget about it:

```bash
komodor auth set-key <your-api-key>
```

Verify the key is active:

```bash
komodor auth show
komodor apikey validate
```

## Usage

```
komodor [command] [subcommand] [flags]
```

### Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` | | API key (overrides env and config file) |
| `--output`, `-o` | `table` | Output format: `table`, `json`, `yaml`, `csv` |
| `--base-url` | | Override the Komodor API base URL |

### Commands

| Command | Description |
|---------|-------------|
| `auth` | Manage authentication credentials |
| `apikey` | Validate the active API key |
| `services` | Search services and fetch their YAML |
| `clusters` | List clusters and user-accessible clusters |
| `jobs` | Search jobs and cronjobs |
| `events` | Search events for a cluster or service |
| `issues` | Search issues for a cluster or service |
| `health` | Manage health risk violations |
| `kubeconfig` | Download kubeconfig from Komodor |
| `audit` | Query audit logs and filter values |
| `users` | Manage users |
| `roles` | Manage RBAC roles |
| `policies` | Manage RBAC policies |
| `actions` | Manage RBAC custom actions |
| `rbac` | Manage user-role assignments |
| `monitors` | Manage realtime monitor configurations |
| `integrations` | Manage Kubernetes cluster integrations |
| `cost` | View cost allocation and right-sizing recommendations |
| `rightsizing` | Manage right-sizing policies |
| `workspaces` | Manage workspaces |
| `klaudia` | Klaudia AI investigation and file management |
| `completion` | Generate shell completion scripts (bash, zsh, fish) |

### Examples

```bash
# List all clusters
komodor clusters list

# Search for services in a specific cluster and namespace
komodor services search --cluster my-cluster --namespace default

# Get the YAML for a specific deployment
komodor services yaml --cluster my-cluster --namespace default --kind deployment --name my-app

# Search issues for a service, output as JSON
komodor issues search --cluster my-cluster --service my-app -o json

# Download a kubeconfig
komodor kubeconfig download --cluster my-cluster

# List users
komodor users list

# Trigger a Klaudia RCA investigation
komodor klaudia rca trigger --cluster my-cluster --service my-app
```

### Shell completion

```bash
# Bash
komodor completion bash > /etc/bash_completion.d/komodor

# Zsh
komodor completion zsh > "${fpath[1]}/_komodor"

# Fish
komodor completion fish > ~/.config/fish/completions/komodor.fish
```

Completion for `--cluster`, `--role`, `--policy`, and `--service-id` flags is dynamic and makes live API calls.

## Contributing

### Prerequisites

- Go 1.23+
- [`golangci-lint`](https://golangci-lint.run/usage/install/)
- [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) v2.4.1 (for regenerating the API client)

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
```

### Project layout

```
cmd/komodor/main.go              # entrypoint
internal/
  auth/auth.go                   # API key resolution
  config/config.go               # ~/.config/komodor/config.yaml read/write
  output/output.go               # Formatter interface (table/json/yaml/csv)
  client/client.gen.go           # generated HTTP client — do not edit
  cmd/
    root.go                      # global flags, auth middleware, context injection
    complete.go                  # dynamic shell completion functions
    helpers.go                   # shared utilities
    testhelpers_test.go          # shared test infrastructure
    *.go                         # one file per resource group
api/
  komodor.yaml                   # original vendored OpenAPI spec
  komodor-clean.yaml             # cleaned spec used for code generation
  oapi-codegen.yaml              # codegen config
```

### Adding a new command

Each resource group lives in its own file (`internal/cmd/<resource>.go`). To add a new command:

1. Create `internal/cmd/<resource>.go` with a parent `newXxxCmd()` grouping subcommands.
2. Each subcommand retrieves its dependencies from context:
   ```go
   c := clientFromCtx(cmd)
   f := formatterFromCtx(cmd)
   ```
3. For list output, implement `output.TableData` on a local struct and pass it to `formatter.Print()`.
4. Register the parent command in `NewRootCmd()` in `root.go`.

### Regenerating the API client

The HTTP client (`internal/client/client.gen.go`) is generated from `api/komodor-clean.yaml`. Never edit it directly.

```bash
make generate
```

If the upstream spec changes, update `api/komodor-clean.yaml` first (strip `x-go-type`/`x-go-type-import` extensions and resolve any duplicate type names), then regenerate.

### Running tests

```bash
# Unit tests (no network, no API key required)
make test

# Integration tests (requires KOMODOR_API_KEY)
KOMODOR_API_KEY=<your-key> make test-integration

# Run a single test by name
go test ./internal/cmd/ -run TestServicesSearch
```

Tests use a `mockClient` injected via `newTestRoot` — no real API calls or config files are needed for unit tests. Any unimplemented method on the mock panics, so unexpected calls surface loudly.

### Linting

```bash
make lint
```

### Before submitting

- Run `make lint` and fix any issues.
- Add or update tests alongside your changes.
- Keep commits focused — one logical change per commit.
