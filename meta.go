package daox

type Column struct {
	ColumnName   string
	IsPrimaryKey bool
}

type TableMeta struct {
	TableName       string
	Columns         []*Column
	PrimaryKey      *Column
	IsAutoIncrement bool
}
