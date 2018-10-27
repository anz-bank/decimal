package decimal

import (
	"fmt"
)

func logicCheck(ok bool, format string, params ...interface{}) {
	if !ok {
		panic(fmt.Errorf("Failed logic check: "+format, params...))
	}
}
