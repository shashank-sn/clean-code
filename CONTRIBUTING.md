# Contributing

Thank you for helping improve Clean Code. This project is open source under the [MIT License](LICENSE).

## Ways to contribute

- **Skills** — improve workflow guidance in `skills/*/SKILL.md`
- **CLI** — extend commands in `cmd/clean-code` and `internal/`
- **Harness** — schemas, adapters, doctrine, and calibration fixtures in `harness/`
- **Docs** — fix or extend `docs/` and `README.md`
- **Tests** — add cases in `tests/` and package `*_test.go` files

## Development setup

```bash
git clone https://github.com/shashank-stitch/clean-code.git
cd clean-code
go test -race ./...
```

Or with npm (requires Node 18+ and Go 1.22+):

```bash
npm install
npm test
```

## Pull requests

1. Fork the repository and create a branch from `codex/initial-release` (or the current default branch).
2. Keep changes focused. Match existing naming and schema conventions.
3. Run `go test -race ./...` before opening a PR.
4. Describe what changed and why. Link an issue when one exists.

## Code standards

- Small, readable functions with clear names.
- Tests for behavior changes; fuzz or table tests for parsers and normalizers.
- No silent weakening of verification gates or policy integrity rules.
- Skills stay host-neutral; doctrine and evidence semantics are canonical.

## Security

Do not open public issues for security vulnerabilities. Email the maintainers through GitHub private reporting if available.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
