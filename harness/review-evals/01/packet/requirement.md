# invoice lookup

An authenticated request may read an invoice only when the invoice belongs to
the tenant in the request claims. A missing or foreign invoice should have the
same not-found result. Keep the handler's caller contract unchanged.
