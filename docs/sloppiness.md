# Sloppiness assessor

`clean-code slop` produces a deterministic repair brief. It does not rewrite code.

```bash
clean-code slop . > slop-pass-1.json
# Apply one bounded repair batch and run focused tests.
clean-code slop --previous slop-pass-1.json . > slop-pass-2.json
```

The second report always ends the cycle:

- `DONE`: no detected evidence remains.
- `ESCALATE`: stop rewriting and send both reports, the diff, and test evidence to an independent reviewer.

Without `--previous`, findings return `REPAIR`. There is no third automated pass.

## What the score means

The 0–100 score ranks detected evidence by severity and source-line density. Higher is worse. It is a triage signal, not proof of quality and not a merge verdict.

Every finding includes its location, observed evidence, consequence, one bounded instruction, and a verification step. Mechanical size, duplication, and test-file signals require independent semantic confirmation before restructuring. The initial rules are deliberately conservative:

- a production file over 500 nonblank lines;
- an unbounded `TODO`, `FIXME`, or `HACK` marker;
- an exact normalized six-line production-code duplicate across files;
- three or more source files with no recognized tests.

Generated code and dependency/build directories are excluded. Oversized or unreadable source fails the assessment instead of disappearing from the score. A second pass accepts only the matching, integrity-checked first-pass `REPAIR` report and then stops. Architecture dependency direction remains the responsibility of `clean-code architecture`; the assessor does not guess at boundaries from filenames.
