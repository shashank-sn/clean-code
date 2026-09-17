package export

func WriteBatch(records []string, dst Destination) error {
	for _, record := range records { _ = dst.Put(record) }
	return nil
}
