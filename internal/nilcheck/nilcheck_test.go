package nilcheck_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwyprof/interactor/v3/internal/nilcheck"
)

type value struct{}

func TestIsNil(t *testing.T) {
	t.Parallel()

	var (
		nilPointer *value
		nilFunc    func()
		nilMap     map[string]int // a nil map can be a working receiver, so it is not "nil" here
	)

	cases := map[string]struct {
		v    any
		want bool
	}{
		"nil interface":  {v: nil, want: true},
		"nil pointer":    {v: nilPointer, want: true},
		"nil func":       {v: nilFunc, want: true},
		"pointer":        {v: &value{}, want: false},
		"func":           {v: func() {}, want: false},
		"struct value":   {v: value{}, want: false},
		"nil map":        {v: nilMap, want: false},
		"non-nil string": {v: "x", want: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, nilcheck.IsNil(tc.v))
		})
	}
}
