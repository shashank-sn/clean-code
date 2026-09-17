# payment creation

Retries carrying the same idempotency key must return the original payment and
create one charge. Concurrent retries are expected. Keep the existing return
value and allow the caller to reset its test store.
