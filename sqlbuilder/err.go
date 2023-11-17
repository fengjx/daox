package sqlbuilder

import (
	"errors"
)

var (
	ErrTableNameRequire = errors.New("[sqlbuilder] tableName requires")
	ErrColumnsRequire   = errors.New("[sqlbuilder] columns requires")
	ErrDeleteMissWhere  = errors.New("[sqlbuilder] delete sql miss where")
)
