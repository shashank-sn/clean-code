# batch export

The export command must report a failure when any destination write fails. It
may retain successfully written records, but callers must never be told that a
fully successful export completed after a partial write.
