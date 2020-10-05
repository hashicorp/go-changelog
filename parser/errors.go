package parser

import "fmt"

type VersionNotFoundErr struct {
	Version string
}

func (e *VersionNotFoundErr) Is(target error) bool {
	tErr, ok := target.(*VersionNotFoundErr)
	if !ok {
		return false
	}
	return tErr.Version == e.Version
}

func (e *VersionNotFoundErr) Error() string {
	return fmt.Sprintf("version %s not found", e.Version)
}
