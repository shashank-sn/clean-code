# Rule authoring

Rules describe one observable concern. Validate them against `harness/schemas/rule.schema.json`.

1. Classify the rule as deterministic, semantic, convention, or architectural.
2. State applicability and evidence. Keep language-specific conventions out of universal doctrine.
3. Give the default severity and false-positive conditions without inventing a universal score.
4. Cite the source or repository decision, then summarize it in original language without copying book passages or code.
5. Add a seeded-defect case and a clean control. A rule is incomplete until it can detect the defect and remain silent on the control.

Correctness, explicit requirements, safety, security, privacy, and data integrity outrank style. A learning proposal can tune quality signals but cannot suppress those protected gates.
