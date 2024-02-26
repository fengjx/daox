package types

import (
	"database/sql"
	"math/big"
	"reflect"
	"time"
)

// enumerates all the database column types
const (
	Bit               = "BIT"
	UnsignedBit       = "UNSIGNED BIT"
	TinyInt           = "TINYINT"
	UnsignedTinyInt   = "UNSIGNED TINYINT"
	SmallInt          = "SMALLINT"
	UnsignedSmallInt  = "UNSIGNED SMALLINT"
	MediumInt         = "MEDIUMINT"
	UnsignedMediumInt = "UNSIGNED MEDIUMINT"
	Int               = "INT"
	UnsignedInt       = "UNSIGNED INT"
	Integer           = "INTEGER"
	BigInt            = "BIGINT"
	UnsignedBigInt    = "UNSIGNED BIGINT"
	Number            = "NUMBER"

	Enum = "ENUM"
	Set  = "SET"

	Char             = "CHAR"
	Varchar          = "VARCHAR"
	VARCHAR2         = "VARCHAR2"
	NChar            = "NCHAR"
	NVarchar         = "NVARCHAR"
	TinyText         = "TINYTEXT"
	Text             = "TEXT"
	NText            = "NTEXT"
	Clob             = "CLOB"
	MediumText       = "MEDIUMTEXT"
	LongText         = "LONGTEXT"
	Uuid             = "UUID"
	UniqueIdentifier = "UNIQUEIDENTIFIER"
	SysName          = "SYSNAME"

	Date          = "DATE"
	DateTime      = "DATETIME"
	SmallDateTime = "SMALLDATETIME"
	Time          = "TIME"
	TimeStamp     = "TIMESTAMP"
	TimeStampz    = "TIMESTAMPZ"
	Year          = "YEAR"

	Decimal    = "DECIMAL"
	Numeric    = "NUMERIC"
	Money      = "MONEY"
	SmallMoney = "SMALLMONEY"

	Real   = "REAL"
	Float  = "FLOAT"
	Double = "DOUBLE"

	Binary     = "BINARY"
	VarBinary  = "VARBINARY"
	TinyBlob   = "TINYBLOB"
	Blob       = "BLOB"
	MediumBlob = "MEDIUMBLOB"
	LongBlob   = "LONGBLOB"
	Bytea      = "BYTEA"

	Bool    = "BOOL"
	Boolean = "BOOLEAN"

	Serial    = "SERIAL"
	BigSerial = "BIGSERIAL"

	Json  = "JSON"
	Jsonb = "JSONB"

	XML   = "XML"
	Array = "ARRAY"
)

// enumerates all types
var (
	IntType   = reflect.TypeOf((*int)(nil)).Elem()
	Int8Type  = reflect.TypeOf((*int8)(nil)).Elem()
	Int16Type = reflect.TypeOf((*int16)(nil)).Elem()
	Int32Type = reflect.TypeOf((*int32)(nil)).Elem()
	Int64Type = reflect.TypeOf((*int64)(nil)).Elem()

	UintType   = reflect.TypeOf((*uint)(nil)).Elem()
	Uint8Type  = reflect.TypeOf((*uint8)(nil)).Elem()
	Uint16Type = reflect.TypeOf((*uint16)(nil)).Elem()
	Uint32Type = reflect.TypeOf((*uint32)(nil)).Elem()
	Uint64Type = reflect.TypeOf((*uint64)(nil)).Elem()

	Float32Type = reflect.TypeOf((*float32)(nil)).Elem()
	Float64Type = reflect.TypeOf((*float64)(nil)).Elem()

	Complex64Type  = reflect.TypeOf((*complex64)(nil)).Elem()
	Complex128Type = reflect.TypeOf((*complex128)(nil)).Elem()

	StringType = reflect.TypeOf((*string)(nil)).Elem()
	BoolType   = reflect.TypeOf((*bool)(nil)).Elem()
	ByteType   = reflect.TypeOf((*byte)(nil)).Elem()
	BytesType  = reflect.SliceOf(ByteType)

	TimeType        = reflect.TypeOf((*time.Time)(nil)).Elem()
	BigFloatType    = reflect.TypeOf((*big.Float)(nil)).Elem()
	NullFloat64Type = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	NullStringType  = reflect.TypeOf((*sql.NullString)(nil)).Elem()
	NullInt32Type   = reflect.TypeOf((*sql.NullInt32)(nil)).Elem()
	NullInt64Type   = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	NullBoolType    = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
)

func SQLType2GoTypeString(sqlType string) string {
	return SQLType2GolangType(sqlType).String()
}

// SQLType2GolangType convert default sql type change to go types
func SQLType2GolangType(sqlType string) reflect.Type {
	switch sqlType {
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, Serial:
		return Int32Type
	case BigInt, BigSerial:
		return Int64Type
	case UnsignedBit, UnsignedTinyInt, UnsignedSmallInt, UnsignedMediumInt, UnsignedInt:
		return UintType
	case UnsignedBigInt:
		return Uint64Type
	case Float, Real:
		return Float32Type
	case Double:
		return Float64Type
	case Char, NChar, Varchar, NVarchar, TinyText, Text, NText, MediumText, LongText, Enum, Set, Uuid, Clob, SysName:
		return StringType
	case TinyBlob, Blob, LongBlob, Bytea, Binary, MediumBlob, VarBinary, UniqueIdentifier:
		return BytesType
	case Bool:
		return BoolType
	case DateTime, Date, Time, TimeStamp, TimeStampz, SmallDateTime, Year:
		return TimeType
	case Decimal, Numeric, Money, SmallMoney:
		return StringType
	default:
		return StringType
	}
}
