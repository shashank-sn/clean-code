package resource

func Serveable(status string) bool { return IsReady(status) }
