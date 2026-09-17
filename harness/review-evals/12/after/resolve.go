package config
import "errors"
var ErrNotFound = errors.New("not found")
type Source interface { Workspace(string) (string, error); Environment(string) (string, error) }
func Resolve(src Source, key string) (string, error) { v, err := src.Workspace(key); if errors.Is(err, ErrNotFound) { return src.Environment(key) }; return v, err }
