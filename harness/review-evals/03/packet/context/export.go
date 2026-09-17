package export

type Destination interface { Put(string) error }
func Run(records []string, dst Destination) error { return WriteBatch(records, dst) }
