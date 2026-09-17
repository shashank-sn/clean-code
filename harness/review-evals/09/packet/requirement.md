# capability lookup

A capability grants access only for its tenant, resource, and token. A token
from another tenant or resource must be denied. Token comparisons should not
leak an early mismatch through ordinary string comparison.
