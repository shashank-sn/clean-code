package rules

type Loader struct{}
func (Loader) Load(raw string) []string { return Parse(raw) }
