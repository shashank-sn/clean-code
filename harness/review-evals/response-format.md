# review response format

For each packet, return a JSON object with:

```json
{
  "case_id": "01",
  "findings": [
    {
      "title": "short problem statement",
      "severity": "blocking|high|medium|low",
      "file": "context/service.go",
      "line": 12,
      "explanation": "why the changed behavior violates the requirement",
      "evidence": "specific caller, input, or execution path"
    }
  ],
  "questions": [],
  "summary": "optional review summary"
}
```

Findings must be grounded in the packet. Do not invent unsupported blocking
issues. A clean control is successful only when `findings` is empty. Omit
unreviewed cases rather than claiming a pass.

The live evaluator records the raw response separately from fixture-validation
results. Oracle files and expected outcomes are not part of the reviewer packet.
