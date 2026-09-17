package resource

func ExampleServeable() {
	// The visible test covers the happy path only; callers still rely on the
	// requirement for non-ready states.
	_ = Serveable("ready")
}
