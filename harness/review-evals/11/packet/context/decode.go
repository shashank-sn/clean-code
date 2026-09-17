package decode

type Decoder struct{}
func (Decoder) All(values []string) ([]int, error) { return ParseAll(values) }
