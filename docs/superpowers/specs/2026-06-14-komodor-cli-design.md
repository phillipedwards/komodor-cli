# Komodor CLI Design

**Date:** 2026-06-14  
**Status:** Implemented

---

## Purpose

`komodor` is a terminal-native CLI that gives SREs and platform engineers full scriptable access to the Komodor platform API without opening a browser. It wraps all ~66 endpoints across 15+ resource groups into a consistent, composable command interface.

---

## Architecture

```
komodor-cli/
├── cmd/komodor/main.go          # entrypoint
├── internal/
│   ├── auth/auth.go             # API key resolution
│   ├── config/config.go         # ~/.config/komodor/config.yaml
│   ├── output/output.go         # Formatter interface + table/json/yaml/csv
│   └── cmd/                     # cobra commands (one file per resource)
│       ├── root.go              # global flags, auth middleware, context injection
│       ├── complete.go          # dynamic completion functions
│       └── *.go                 # resource command files
├── api/
│   ├── komodor.yaml             # original OpenAPI spec (vendored)
│   ├── komodor-clean.yaml       # cleaned spec (x-go-type stripped, name conflicts resolved)
│   └── oapi-codegen.yaml        # codegen config
└── internal/client/client.gen.go  # generated HTTP client (DO NOT EDIT, 22k lines)
```

### Request flow

```
cobra PersistentPreRunE
  → auth.Resolve(flag > KOMODOR_API_KEY > config)
  → client.NewClientWithResponses(baseURL, x-api-key header)
  → store (client, formatter) in cobra context
    → command RunE
        → clientFromCtx(cmd)      # returns ClientWithResponsesInterface
        → formatterFromCtx(cmd)   # returns Formatter
        → call generated client method
        → formatter.Print(result)
```

---

## Key Design Decisions

### Generated client, hand-written commands

The API client is generated from `api/komodor-clean.yaml` using `oapi-codegen v2.4.1`. This gives fully typed request/response structs with no hand-maintained HTTP plumbing. All 66 endpoints are covered. Cobra commands act as a thin UX layer on top.

The spec required cleaning before codegen could succeed:
- Stripped 58 `x-go-type`/`x-go-type-import` extensions referencing private Komodor packages
- Replaced null schemas (`{}`) where properties had only `x-go-type` content
- Resolved duplicate typenames (`MonitorType`, `Statement1`/`PolicyStatement`, `Policy1`/`PolicyV2`)

### Context-based dependency injection

The API client and output formatter are stored in cobra's context by `PersistentPreRunE` and retrieved in each command via `clientFromCtx` / `formatterFromCtx`. This keeps command functions side-effect-free and trivially testable — tests replace `PersistentPreRunE` to inject mocks.

`clientFromCtx` returns `client.ClientWithResponsesInterface` (the generated interface), not the concrete `*ClientWithResponses`. This decouples commands from the concrete type and enables mock injection without external libraries.

### Auth priority

```
--api-key flag  >  KOMODOR_API_KEY env var  >  ~/.config/komodor/config.yaml
```

Auth-exempt commands (no key required): `auth set-key`, `auth show`, `completion`, `help`, `__complete`, `__completeNoDesc`.

### Output formatting

The `Formatter` interface has a single method: `Print(data any) error`.

Commands that return lists implement `output.TableData` (`Headers() []string`, `Rows() [][]string`) for table and CSV rendering. All other types fall back to JSON. Available formats: `table` (default), `json`, `yaml`, `csv`.

### Dynamic shell completion

Five flags get live completions fetched from the API:

| Flag | Source endpoint |
|---|---|
| `--cluster` | `GET /api/v2/clusters` |
| `--role` | `GET /api/v2/rbac/roles` |
| `--policy` | `GET /api/v2/rbac/policies` |
| `--service-id` | `POST /api/v2/services/search` |

Completion functions build their own API client from env/config (flags are not parsed during completion). Any auth or network failure silently returns empty — tab completion must never break the shell.

---

## Command Reference

```
komodor auth set-key <key>
komodor auth show

komodor apikey validate

komodor completion bash|zsh|fish

komodor clusters (cluster) list [--name]
komodor clusters user-clusters

komodor services (svc) search [--cluster] [--namespace] [--page-size]
komodor services yaml   --cluster --namespace --kind --name

komodor jobs search [--cluster] [--namespace] [--page] [--page-size]

komodor events (event) search         --service-id [--from] [--to]
komodor events search-cluster         --cluster    [--from] [--to]
komodor events create                 --title [--service-id] [--message]

komodor issues (issue) search         --service-id --from --to
komodor issues search-cluster         --cluster    --from --to

komodor health risks list  [--cluster] [--severity]
komodor health risks get   <id>
komodor health risks update <id> --status

komodor kubeconfig get [--cluster] [--connection]

komodor audit logs    [--cluster] [--from] [--to] [--action] [--user] ...
komodor audit filters [--from] [--to]

komodor users (user) list
komodor users get   <id-or-email>
komodor users create --email [--name]
komodor users update <id-or-email> [--name] [--role]
komodor users delete <id-or-email>
komodor users effective-permissions [--id-or-email]

komodor roles (role) list
komodor roles get    <id-or-name>
komodor roles create --name [--policy]
komodor roles update <id-or-name> [--name] [--policy]
komodor roles delete <id-or-name>
komodor roles attach-policy --role --policy
komodor roles detach-policy --role --policy

komodor policies (policy) list
komodor policies get    <id-or-name>
komodor policies create --name
komodor policies update <id-or-name> [--name]
komodor policies delete <id-or-name>

komodor actions (action) list
komodor actions get    <id>
komodor actions create --name --verb --resource
komodor actions update <id> [--name] [--verb] [--resource]
komodor actions delete <id>

komodor rbac users attach --user --role
komodor rbac users update --user --role [--expiration]
komodor rbac users detach --user --role

komodor monitors (monitor) list
komodor monitors get    <id>
komodor monitors create --name --type --cluster
komodor monitors update <id> [flags]
komodor monitors delete <id>

komodor integrations (integration) k8s create --cluster
komodor integrations k8s get    <cluster-name>
komodor integrations k8s delete <cluster-name>

komodor cost allocation    --time-frame --group-by [--cluster]
komodor cost right-sizing services   [--cluster] [--namespace] [--strategy]
komodor cost right-sizing containers --cluster --namespace --kind --name

komodor right-sizing-policies (rsp) list
komodor right-sizing-policies get      <id>
komodor right-sizing-policies defaults
komodor right-sizing-policies create   --name
komodor right-sizing-policies update   <id> [--name]
komodor right-sizing-policies delete   <id> [--force]

komodor workspaces (ws) list [--page] [--page-size]
komodor workspaces get    <id>
komodor workspaces create --name [--description]
komodor workspaces update <id> --name [--description]
komodor workspaces delete <id>

komodor klaudia (ai) rca trigger --cluster --name --namespace --kind [--issue-id]
komodor klaudia rca get <session-id>
komodor klaudia files list   <type>
komodor klaudia files get    <type> <file-id>
komodor klaudia files upload <type> --file
komodor klaudia files update <type> <file-id> --file
komodor klaudia files delete <type> --file-id
```

---

## Error Handling

- All errors to stderr; cobra handles exit codes
- No API key → `"no API key configured — set KOMODOR_API_KEY or run 'komodor auth set-key <key>'"`
- API 4xx/5xx → `"API error <status>: <body>"` via `apiError(statusCode, body)`

---

## Build & Development

```bash
make generate   # regenerate internal/client/client.gen.go from api/komodor-clean.yaml
make build      # produces ./bin/komodor
make test       # unit tests (no network required)
make test-integration  # requires KOMODOR_API_KEY
```

Go version: 1.23.1  
Key dependencies: `cobra`, `oapi-codegen v2.4.1`, `tablewriter v1.1.4`, `gopkg.in/yaml.v3`
