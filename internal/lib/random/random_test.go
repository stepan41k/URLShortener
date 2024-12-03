package random_test

import (
	"testing"

	"github.com/stepan41k/FullRestAPI/internal/lib/random"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "size = 1",
			size: 1,
		},
		{
			name: "size = 5",
			size: 5,
		},
		{
			name: "size = 10",
			size: 10,
		},
		{
			name: "size = 20",
			size: 20,
		},
	}

	for _, v := range tests {
		t.Logf("%s: %s", v.name, random.NewRandomString(v.size))
	}

}