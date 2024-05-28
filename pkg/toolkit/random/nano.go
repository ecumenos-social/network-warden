package random

import (
	goNanoID "github.com/matoous/go-nanoid"
)

func GenNanoString(l int) (string, error) {
	return goNanoID.ID(l)
}
