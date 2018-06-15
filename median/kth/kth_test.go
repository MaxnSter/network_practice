package kth

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestFindKth(t *testing.T) {
	a := []int{1, 2, 2, 4, 5, 6, 7}
	ak, _ := FindKth(func(guess int) (smaller, same int) {
		for _, v := range a {
			if v < guess {
				smaller++
			}

			if v == guess {
				same++
			}
		}
		return
	}, 4, len(a), 1, 7)

	assert.Equal(t, ak, 4)
}

func TestFindKth2(t *testing.T) {
	a := []int{1, 2, 2, 4, 5, 6}
	ak, finded := FindKth(func(guess int) (smaller, same int) {
		for _, v := range a {
			if v < guess {
				smaller++
			}

			if v == guess {
				same++
			}
		}
		return
	}, 3, len(a), 1, 7)

	assert.Equal(t, finded, true)
	assert.Equal(t, ak, 2)
}
