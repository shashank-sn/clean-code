package command

func Build(child []string, raw string) []string { return append(child, DecodeFlags(raw)...)}
