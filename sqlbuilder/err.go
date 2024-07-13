package sqlbuilder

import (
	"errors"
)

var (
	ErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
	ErrUpdateMissWhere  = errors.New("[sqlbuilder] where express requires with update")
	ErrColumnsRequire   = errors.New("[sqlbuilder] columns requires")
	ErrDeleteMissWhere  = errors.New("[sqlbuilder] delete sql miss where")
	ErrExecerNotSet     = errors.New("[sqlbuilder] execer not set")
	ErrQueryerNotSet    = errors.New("[sqlbuilder] queryer not set")
)
