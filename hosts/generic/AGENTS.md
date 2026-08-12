# Clean Code workflow

Use the repository's own commands and conventions as the source of truth.

1. Run `clean-code discover <repo>` before proposing checks.
2. Define requirements, acceptance examples, and dependency boundaries before implementation.
3. Keep implementation, acceptance testing, and review responsibilities separate when the host allows it.
4. Run `clean-code verify <repo>` before declaring completion.
5. Report `NOT_AVAILABLE`, `NOT_CONFIGURED`, `NOT_RUN`, and `ERROR` exactly. Never convert them to `PASS`.
6. Require concrete evidence for review findings. A correct review may return zero findings.
7. Record human spot checks for requirements, acceptance examples, UI/QA procedures, and sampled code.
