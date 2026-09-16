---
name: clean-prune
description: Delete dead, unreachable, and leftover code from the change set, with a recorded reference search behind every deletion. Use after implementation and simplification, before final verification and review.
---

# Clean Prune

Ship a change set with no dead weight. Delete code nothing uses.

## Workflow

1. Scope to the branch diff, the files it touches, and the immediate dependency edges of every symbol the change added, changed, or superseded.
2. Collect candidates with detectors the repository already declares (`go vet`, `staticcheck -unused`, `deadcode`, `knip`, `ts-prune`, `vulture`, `unused`, `cargo udeps`, `cargo machete`, coverage reports). Never install a detector, dependency, or tool to manufacture evidence.
3. Treat detector output as a candidate list, not proof. Confirm each candidate by searching the repository for references before deleting: direct use, re-exports, string or configuration lookups, reflection, dependency injection, framework and route conventions, templates, and build files.
4. Classify each candidate as REMOVE, KEEP with a reason, or UNRESOLVED. Record the path, symbol, and exact search behind the classification.
5. Remove the REMOVE set outright. Delete unused functions, types, fields, parameters, imports, variables, branches, files, fixtures, assets, and flags; delete superseded implementations instead of parking them beside their replacement. No commented-out blocks, no `// removed` markers, no re-export shims, no renamed placeholder variables.
6. Re-run the affected tests and declared checks. If a removal changes behavior, restore it and record it as a separate bounded fix rather than folding it into cleanup.
7. Hand off to `clean-verify` with the pruned revision, then to `clean-review`.

## Candidate categories

- Unreferenced functions, types, constants, and package-level declarations.
- Unreachable statements, dead branches, and conditions that can no longer be true.
- Unused imports, variables, parameters, fields, and configuration keys.
- Leftover scaffolding: debug prints, temporary flags, hardcoded probe values, empty stub bodies, and feature flags with no live reader.
- Superseded implementations and duplicated helpers the change replaced.
- Orphaned files, fixtures, assets, and test helpers nothing imports or runs.
- Unused dependencies and stale build or manifest entries.
- Commented-out code and backward-compatibility shims with no remaining consumer.

## Preserve

- The published contract surface: exported types, CLI flags, wire formats, schema fields, and public routes keep their meaning unless the requirements changed them.
- Structure pins from the plan or `session-settled` decisions, including deliberate duplication.
- Generated, vendored, or migration code, and fixtures that exist for a reason outside the current call graph.
- Code reached only through reflection, dynamic import, string lookup, dependency injection, framework convention, templates, or configuration.
- Deprecations that document a consumer and a removal condition.

## Evidence

- List every removal with its path, symbol, and the reference search that justified it.
- List every detector candidate kept, with the reason it stays.
- Report detector state as `NOT_AVAILABLE` (no detector for the language), `NOT_CONFIGURED` (none declared), `NOT_RUN` (skipped), or `ERROR` (failed). Never present an unconfirmed detector count as a pass.

## Gates

- Block ship while an UNRESOLVED candidate remains in the change set.
- Block any removal that has no recorded reference search.
- Do not invent a dead-code score, threshold, or percentage. Deletion evidence stays separate from coverage, mutation, duplication, complexity, and `clean-verify` status.
- Keep behavior identical. Pruning is not a behavior change and does not replace `clean-test`.

## Safety

- Search before deleting; never delete on a name heuristic alone.
- Keep deletions inside the reviewed scope. Flag dead code outside the change set instead of sweeping it.
- Re-run affected tests when a removal touches a tested or executed path.
- Report unavailable tooling as a gap rather than skipping it silently.
