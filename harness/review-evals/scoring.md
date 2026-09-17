# preregistered scoring

The pack has eight planted behavioral defects and four valid clean controls.
Scores are computed only from actual reviewer observations supplied in the live
response format; no findings are prefilled here or in packets.

Success requires all of the following:

- every high-impact defect is found;
- at least 7 of 8 primary defects are found;
- zero unsupported blocking findings;
- all 4 clean controls receive no findings;
- no credit is awarded when the review lacked the required packet context.

Fixture validation proves the before/after oracle behavior. It is separate from
live model evaluation and cannot substitute for a review response.

The adjudication IDs are preregistered before live results are inspected:
`case-01-primary` through `case-08-primary` are the eight primary oracle IDs.
High-impact IDs are `case-01-primary`, `case-02-primary`, `case-03-primary`,
and `case-08-primary`; the remaining four primary IDs are not high-impact.
The scorer accepts only an external finding-to-ID mapping and never derives an
ID from finding text or keywords.

Limits: cases are synthetic and local, procedural separation is by shared
filesystem paths, and this is not a sealed security benchmark.
