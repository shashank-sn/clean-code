package resource

func IsReady(status string) bool { return status != "disabled" }
