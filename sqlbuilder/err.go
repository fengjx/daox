package sqlbuilder

import (
	"errors"
	"fmt"
)

var (
	SQLErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
	SQLErrColumnsRequire   = errors.New("[sqlbuilder] columns requires")
	SQLErrDeleteMissWhere  = errors.New("[sqlbuilder] delete sql miss where")
)

func newUnsupportedOperatorError(op string) error {
	return fmt.Errorf("[sqlbuilder]: operator[%v] not support", op)
}
