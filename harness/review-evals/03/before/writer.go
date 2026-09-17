package export

type Destination interface { Put(string) error }
func WriteBatch(records []string, dst Destination) error {
	for _, record := range records {
		if err := dst.Put(record); err != nil { return err }
	}
	return nil
}
