package validator

import (
	"testing"

	"scadrialapi.abdulmoiz.net/internal/assert"
)

func TestValidator(t *testing.T) {
	validatorTest := New()
	t.Run("Valid validator", func(t *testing.T) {
		res := validatorTest.Valid()
		assert.Equal(t, res, true)

	})
	t.Run("Add Error", func(t *testing.T) {
		validatorTest.AddError("abc","xyz")
		res := validatorTest.Valid()
		assert.Equal(t, res, false)
	})
	t.Run("Unique values", func(t *testing.T) {
		tests := [] struct {
			name string
			arr [] int
			want bool
		}{
			{
			name: "Unique array of int",
			arr: []int{1,2,3,4},
			want: true,
			},
			{
			name: "Non unique array of int",
			arr: []int{1,2,4,4},
			want: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				res := Unique(tt.arr)
				assert.Equal(t, res, tt.want)

			})
		}

	})

	
}