# Architecture policy

The architecture checker compares a repository-owned component policy with a dependency graph produced by the repository's chosen language or build tool.

## Policy

Components declare path patterns, permitted component dependencies, and optional public surfaces. `/**` matches a directory recursively; other patterns follow the host operating system's file matching rules.

```json
{
  "schema_version": "1.0.0",
  "exclude": ["generated/**"],
  "components": [
    {
      "id": "core",
      "paths": ["core/**"],
      "public": ["core/api/**"]
    },
    {
      "id": "delivery",
      "paths": ["delivery/**"],
      "may_depend_on": ["core"]
    }
  ],
  "exceptions": [
    {
      "from": "core/legacy.go",
      "to": "delivery/legacy.go",
      "reason": "remove after the transport migration"
    }
  ]
}
```

An omitted `may_depend_on` list allows same-component edges only. An omitted `public` list leaves every file in that component reachable from permitted components. `allow_cycles` defaults to false.

## Dependency graph

The core accepts a language-neutral edge list:

```json
{
  "schema_version": "1.0.0",
  "edges": [
    {"from": "delivery/http.go", "to": "core/api/usecase.go"}
  ]
}
```

Paths must be repository-relative. The graph producer owns language resolution, generated edges, and completeness.

## Run

```bash
clean-code architecture \
  --policy .clean-code-architecture.json \
  --graph build/dependency-graph.json
```

The command exits zero for `PASS` and one for `FAIL` or invalid input. Violations distinguish forbidden dependencies, access outside a public surface, undeclared paths, ambiguous component ownership, and component cycles. Cycle results include the dependency path.

## Limits

Path rules make approved dependency direction executable. Questions about component names, responsibilities, and use-case boundaries remain part of semantic design review and acceptance evidence.
