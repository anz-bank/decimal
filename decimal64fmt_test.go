package decimal

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimal64String(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		expected := strconv.Itoa(int(i))
		d := NewDecimal64FromInt64(i)
		actual := d.String()
		require.Equal(t, expected, actual)
	}

	for f := 1; f < 100; f += 2 {
		fraction := NewDecimal64FromInt64(int64(f)).Quo(NewDecimal64FromInt64(100))
		for i := int64(0); i <= 100; i++ {
			expected := strconv.Itoa(int(i)) + "." + strconv.Itoa(100 + f)[1:3]
			d := NewDecimal64FromInt64(i).Add(fraction)
			actual := d.String()
			require.Equal(t, expected, actual)
		}
		for i := int64(-100); i < 0; i++ {
			expected := strconv.Itoa(int(i)) + "." + strconv.Itoa(100 + f)[1:3]
			d := NewDecimal64FromInt64(i).Sub(fraction)
			actual := d.String()
			require.Equal(t, expected, actual)
		}
	}
}

func TestDecimal64Format(t *testing.T) {
	for i := int64(0); i <= 1000; i++ {
		expected := strconv.FormatInt(i, 10)
		d := NewDecimal64FromInt64(i)
		actual := fmt.Sprintf("%v", d)
		require.Equal(t, expected, actual)
	}
}
