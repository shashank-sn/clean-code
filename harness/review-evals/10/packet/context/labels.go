package labels

type Store struct{ Values []string }
func (s *Store) Replace(values []string) { s.Values = Normalize(values) }
