package random

import (
	"math/rand"
)

var numericRunes = []rune("0123456789")

func GenNumericString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = numericRunes[rand.Intn(len(numericRunes))]
	}
	return string(b)
}
