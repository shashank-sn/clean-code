# Full-flow benchmark task: slug normalizer

## Requirement R1

Add `NormalizeSlug(input string) string` that produces URL-safe slugs.

## Behavior

1. Trim leading and trailing whitespace from the input.
2. Lowercase ASCII letters.
3. Replace spaces and underscores with a single hyphen.
4. Remove characters that are not `a-z`, `0-9`, or hyphen.
5. Collapse consecutive hyphens to one hyphen.
6. Trim leading and trailing hyphens from the result.
7. Return empty string when the input is empty or produces no valid characters.

## Examples

| Input | Output |
| --- | --- |
| `Hello World` | `hello-world` |
| `  Foo__Bar!!  ` | `foo-bar` |
| `already-clean` | `already-clean` |
| `---` | `` |
| `` | `` |

## Constraints

- Pure function, no globals, no I/O.
- Table-driven unit tests required for CC workflow evidence.

This task is intentionally small so Compound Engineering and Clean Code workflows can complete it in one session while still exposing differences in structure, tests, and review discipline.
