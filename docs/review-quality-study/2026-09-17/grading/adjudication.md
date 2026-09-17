# adjudication

All 12 cases in amber, juniper, and river had complete source-packet reviews; each assigned case was present in its paired raw run. Every raw finding was retained and matched to the corresponding fixed primary oracle for cases 01-08. The four controls (09-12) had no findings in any condition.

The eight planted defects were concrete causal matches: tenant isolation (01), concurrent idempotency (02), destination error propagation (03), comma delimiter decoding (04), ready-state allowlisting (05), escaped separator parsing (06), strict expiry (07), and shutdown waiting (08). No unsupported or ambiguous findings were observed; no extra findings were present.

The protocol probe was incomplete in every condition because the packet intentionally omitted implementation and caller context. Each raw probe asked for the missing context, made no defect assertion, and withheld approval. Probe results are recorded separately and do not receive primary credit.

Counts per condition: 12 complete case reviews; 8 supported primary findings; 0 control findings; 0 unsupported findings; 0 ambiguous findings; 1 incomplete protocol probe.
