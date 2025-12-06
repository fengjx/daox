package gen

// Config config file
type Config struct {
	DS     *DS           `yaml:"ds"`
	Target ReverseTarget `yaml:"target"`
}

type DS struct {
	Type string `yaml:"type"`
	Dsn  string `yaml:"dsn"`
}

type Var map[string]string

type RelationConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"` // o2m, o2o, m2o
	ForeignKey string `yaml:"foreign_key"`
	Table      string `yaml:"table"`
	RefKey     string `yaml:"ref_key"`
}

type TableConfig struct {
	ShortName       string           `yaml:"short-name"`
	TargetDir       string           `yaml:"target-dir"` // 可选，生成文件路径，默认 ./
	Var             Var              `yaml:"var"`
	Relations       []RelationConfig `yaml:"relations"`
	ShortNameLower  string           `yaml:"-"`
	TableNameLower  string           `yaml:"-"`
	DaoGenGoImports []string         `yaml:"-"`
}

type ReverseTarget struct {
	Custom Custom                 `yaml:"custom"` // 自定义参数
	Tables map[string]TableConfig `yaml:"tables"` // 生成的数据库表名和对应的自定义变量
}

type Custom struct {
	Gomod       string `yaml:"gomod"`        // 必须，项目 go mod module 配置
	TargetDir   string `yaml:"target-dir"`   // 可选，生成文件路径，默认 ./
	TemplateDir string `yaml:"template-dir"` // 可选，模板路径
	UseAdmin    bool   `yaml:"use-admin"`    // 可选，是否生成管理后台接口和页面，对所有表生效
	Var         Var    `yaml:"exy"`          // 可选，自定义变量，对所有表生效
}

// Table represents a database table
type Table struct {
	Name           string
	StructName     string
	Columns        []Column
	PrimaryKey     Column
	AutoIncrement  bool
	Comment        string
	StoreEngine    string
	GoImports      []string
	CreateTableSQL string // SHOW CREATE TABLE 的结果
}

type Column struct {
	TableName    string
	Name         string
	SQLType      string
	Comment      string
	IsPrimaryKey bool
	IsTimeType   bool
	DefaultValue string
	Extra        string
}
