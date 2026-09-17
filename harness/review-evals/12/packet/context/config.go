package config

import "errors"
var ErrNotFound = errors.New("not found")
type Source interface { Workspace(string) (string, error); Environment(string) (string, error) }
