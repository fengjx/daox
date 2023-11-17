package main

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/fengjx/daox/utils"
)

//go:embed template/*
var embedFS embed.FS

func main() {
	app := &cli.App{
		Name:        "code-gen",
		Description: "create template file from database",
		Version:     "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "f",
				Usage:    "config file path",
				Required: true,
			},
		},
		Action: run,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	configFile := ctx.String("f")
	bs, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	config := &Config{}
	if err = yaml.Unmarshal(bs, config); err != nil {
		return err
	}
	if config.DS.Type != "mysql" {
		fmt.Println("only support mysql now")
		return nil
	}
	dsnCfg, err := mysql.ParseDSN(config.DS.Dsn)
	if err != nil {
		log.Fatal(err)
	}
	db := sqlx.MustOpen(config.DS.Type, config.DS.Dsn)
	db.Mapper = reflectx.NewMapperFunc("db", strings.ToTitle)
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	for _, tableName := range config.Target.Tables {
		table := loadTableMeta(db, dsnCfg.DBName, tableName)
		fmt.Println(table.Name, table.Comment)
		gen(config, table)
	}
	return nil
}

func loadTableMeta(db *sqlx.DB, dbName, tableName string) *Table {
	args := []interface{}{dbName, tableName}
	querySQL := "SELECT `TABLE_NAME`, `ENGINE`, `AUTO_INCREMENT`, `TABLE_COMMENT` from" +
		" `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=? AND TABLE_NAME = ?" +
		" AND (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB')"

	rows, err := db.Query(querySQL, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	table := new(Table)
	for rows.Next() {
		var name, engine string
		var comment *string
		var autoIncr *int
		err = rows.Scan(&name, &engine, &autoIncr, &comment)
		if err != nil {
			log.Fatal(err)
		}
		table.Name = name
		table.StructName = utils.GonicCase(name)
		if comment != nil {
			table.Comment = *comment
		}
		table.StoreEngine = engine
		if autoIncr != nil {
			table.AutoIncrement = true
		}
	}
	if rows.Err() != nil {
		log.Fatal(err)
	}
	columns, primaryKey := loadColumnMeta(db, dbName, tableName)
	table.Columns = columns
	table.PrimaryKey = primaryKey
	table.GoImports = GenGoImports(table.Columns)
	return table
}

// loadColumnMeta
// []*Column table column meta
// *Column PrimaryKey column
func loadColumnMeta(db *sqlx.DB, dbName, tableName string) ([]Column, Column) {
	args := []interface{}{dbName, tableName}
	querySQL := "SELECT column_name, column_type, column_comment, column_key FROM information_schema.columns " +
		"WHERE table_schema = ? AND table_name = ? ORDER BY ORDINAL_POSITION"
	rows, err := db.Query(querySQL, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var columns []Column
	var primaryKey Column
	for rows.Next() {
		var columnName string
		var columnType string
		var columnComment string
		var columnKey string
		err = rows.Scan(&columnName, &columnType, &columnComment, &columnKey)
		if err != nil {
			log.Fatal(err)
		}
		col := Column{}
		col.Name = strings.Trim(columnName, "` ")
		col.Comment = columnComment

		fields := strings.Fields(columnType)
		columnType = fields[0]
		cts := strings.Split(columnType, "(")
		colName := cts[0]
		// Remove the /* mariadb-5.3 */ suffix from coltypes
		colName = strings.TrimSuffix(colName, "/* mariadb-5.3 */")
		col.SQLType = strings.ToUpper(colName)

		if columnKey == "PRI" {
			col.IsPrimaryKey = true
			primaryKey = col
		}
		columns = append(columns, col)
	}
	return columns, primaryKey
}

func gen(config *Config, table *Table) {
	dir := "template/default"
	isEmbed := true
	if config.Target.Custom.TemplateDir != "" {
		dir = config.Target.Custom.TemplateDir
		isEmbed = false
	}
	entries, err := ReadDir(dir, isEmbed)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	attr := map[string]interface{}{
		"Var":     config.Target.Custom.Var,
		"TagName": config.Target.Custom.TagName,
		"Table":   table,
	}
	out := filepath.Join(config.Target.Custom.OutDir)
	render(isEmbed, filepath.Join(dir), "", entries, out, attr)
}

// render 递归生成文件
func render(isEmbed bool, basePath string, parent string, entries []os.DirEntry, outDir string, attr map[string]interface{}) {
	if parent == "" {
		parent = basePath
	}
	for _, entry := range entries {
		path := filepath.Join(parent, entry.Name())
		if entry.IsDir() {
			children, err := ReadDir(filepath.Join(parent, entry.Name()), isEmbed)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			render(isEmbed, basePath, path, children, outDir, attr)
			continue
		}
		targetDir := filepath.Join(outDir, strings.ReplaceAll(parent, basePath, ""))
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		if !strings.HasSuffix(entry.Name(), ".tmpl") {
			targetFile := filepath.Join(targetDir, entry.Name())
			fmt.Println(targetFile)
			// 其他不需要渲染的文件直接复制
			err = utils.CopyFile(path, targetFile)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			continue
		}
		filenameBytes, err := parse(strings.ReplaceAll(entry.Name(), ".tmpl", ""), attr)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		targetFile := filepath.Join(targetDir, string(filenameBytes))
		fmt.Println(targetFile)
		bs, err := ReadFile(path, isEmbed)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		newbytes, err := parse(string(bs), attr)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = os.WriteFile(targetFile, newbytes, 0600)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func parse(text string, attr map[string]interface{}) ([]byte, error) {
	funcMap := template.FuncMap{
		"FirstUpper":           utils.FirstUpper,
		"FirstLower":           utils.FirstLower,
		"SnakeCase":            utils.SnakeCase,
		"TitleCase":            utils.TitleCase,
		"GonicCase":            utils.GonicCase,
		"LineString":           utils.LineString,
		"SQLType2GoTypeString": SQLType2GoTypeString,
	}
	t, err := template.New("").Funcs(funcMap).Parse(text)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString("")
	if err := t.Execute(buf, attr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Config config file
type Config struct {
	DS     *DS
	Target *ReverseTarget
}

type DS struct {
	Type string
	Dsn  string
}

type ReverseTarget struct {
	Custom *Custom
	Tables []string
}

type Custom struct {
	TemplateDir string            `yaml:"template-dir"`
	OutDir      string            `yaml:"out-dir"`
	Var         map[string]string `yaml:"var"`
	TagName     string            `yaml:"tag-name"`
}

// Table represents a database table
type Table struct {
	Name          string
	StructName    string
	Columns       []Column
	PrimaryKey    Column
	AutoIncrement bool
	Comment       string
	StoreEngine   string
	GoImports     []string
}

type Column struct {
	TableName    string
	Name         string
	SQLType      string
	Comment      string
	IsPrimaryKey bool
}

// enumerates all the database column types
var (
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

func GenGoImports(cols []Column) []string {
	imports := make(map[string]string)
	results := make([]string, 0)
	for _, col := range cols {
		if SQLType2GolangType(col.SQLType) == TimeType {
			if _, ok := imports["time"]; !ok {
				imports["time"] = "time"
				results = append(results, "time")
			}
		}
	}
	return results
}

func ReadDir(name string, isEmbed bool) ([]fs.DirEntry, error) {
	if isEmbed {
		return embedFS.ReadDir(name)
	}
	return os.ReadDir(name)
}

func ReadFile(name string, isEmbed bool) ([]byte, error) {
	if isEmbed {
		return embedFS.ReadFile(name)
	}
	return os.ReadFile(name)
}
