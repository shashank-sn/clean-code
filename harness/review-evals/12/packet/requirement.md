# configuration fallback

Use the workspace configuration when present. Fall back to the process
environment only when the workspace lookup returns `ErrNotFound`; preserve any
other lookup error for the caller.
