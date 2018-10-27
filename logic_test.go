package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogicCheck(t *testing.T) {
	require.NotPanics(t, func() { logicCheck(true, "") })
	require.Panics(t, func() { logicCheck(false, "") })
}
