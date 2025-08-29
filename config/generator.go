package generator

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config 生成器配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Output   string
	Package  string
	Tables   string
	// 新增配置项
	GenerateRouter  bool
	GenerateService bool
	RouterOutput    string
	ServiceOutput   string
	// 生成代码所需的导入路径
	// ModelImportPath 指向生成的模型包导入路径（用于在 Router/Service 中引用模型）。
	// 例如: "github.com/your/app/internal/models"
	ModelImportPath   string
	// ServiceImportPath 指向服务层包的导入路径（用于在 Router 中引用服务）。
	// 例如: "github.com/your/app/internal/services"
	ServiceImportPath string
	// StorageImportPath 指向存储层包的导入路径根（用于在 Service 中引用存储实现，如 DAO/Repository）。
	// 例如: "github.com/your/app/internal/storage"
	StorageImportPath string
}

// TableInfo 表信息
type TableInfo struct {
	Name        string
	Comment     string
	Columns     []ColumnInfo
	PrimaryKeys []string
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string
	Type         string
	Comment      string
	IsNullable   bool
	IsPrimaryKey bool
	IsAutoIncr   bool
	DefaultValue string
	GoType       string
	GoTag        string
}

// Generator 代码生成器
type Generator struct {
	config *Config
	db     *sql.DB
}

// NewGenerator 创建生成器
func NewGenerator(config *Config) *Generator {
	// 设置默认输出路径
	if config.RouterOutput == "" {
		config.RouterOutput = filepath.Join(config.Output, "../router")
	}
	if config.ServiceOutput == "" {
		config.ServiceOutput = filepath.Join(config.Output, "../services")
	}
	
	return &Generator{
		config: config,
	}
}

// Generate 生成代码
func (g *Generator) Generate() error {
	// 连接数据库
	if err := g.connectDB(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer g.db.Close()

	// 获取表信息
	tables, err := g.getTables()
	if err != nil {
		return fmt.Errorf("获取表信息失败: %w", err)
	}

	// 创建输出目录
	if err := os.MkdirAll(g.config.Output, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成基础模型文件
	if err := g.generateBaseModel(); err != nil {
		return fmt.Errorf("生成基础模型失败: %w", err)
	}

	// 生成每个表的模型文件
	for _, table := range tables {
		if err := g.generateTableModel(table); err != nil {
			log.Printf("生成表 %s 的模型失败: %v", table.Name, err)
			continue
		}
		fmt.Printf("生成表 %s 的模型成功\n", table.Name)
	}

	// 生成 Service 代码
	if g.config.GenerateService {
		if err := g.generateServices(tables); err != nil {
			return fmt.Errorf("生成 Service 代码失败: %w", err)
		}
	}

	// 生成 Router 代码
	if g.config.GenerateRouter {
		if err := g.generateRouters(tables); err != nil {
			return fmt.Errorf("生成 Router 代码失败: %w", err)
		}
	}

	return nil
}

// connectDB 连接数据库
func (g *Generator) connectDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		g.config.User,
		g.config.Password,
		g.config.Host,
		g.config.Port,
		g.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	g.db = db
	return nil
}

// getTables 获取表信息
func (g *Generator) getTables() ([]TableInfo, error) {
	var tables []TableInfo

	// 构建查询条件
	whereClause := ""
	if g.config.Tables != "" {
		tableNames := strings.Split(g.config.Tables, ",")
		placeholders := make([]string, len(tableNames))
		for i := range tableNames {
			placeholders[i] = "?"
		}
		whereClause = fmt.Sprintf("WHERE TABLE_NAME IN (%s)", strings.Join(placeholders, ","))
	}

	query := fmt.Sprintf(`
		SELECT 
			TABLE_NAME,
			TABLE_COMMENT
		FROM 
			INFORMATION_SCHEMA.TABLES 
		%s
		AND TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`, whereClause)

	var args []interface{}
	if g.config.Tables != "" {
		tableNames := strings.Split(g.config.Tables, ",")
		for _, name := range tableNames {
			args = append(args, strings.TrimSpace(name))
		}
	}
	args = append(args, g.config.Database)

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Comment); err != nil {
			return nil, err
		}

		// 获取列信息
		columns, err := g.getColumns(table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		// 获取主键信息
		primaryKeys, err := g.getPrimaryKeys(table.Name)
		if err != nil {
			return nil, err
		}
		table.PrimaryKeys = primaryKeys

		tables = append(tables, table)
	}

	return tables, nil
}

// getColumns 获取列信息
func (g *Generator) getColumns(tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			COLUMN_COMMENT,
			IS_NULLABLE,
			COLUMN_KEY,
			EXTRA,
			COLUMN_DEFAULT
		FROM 
			INFORMATION_SCHEMA.COLUMNS 
		WHERE 
			TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY 
			ORDINAL_POSITION
	`

	rows, err := g.db.Query(query, g.config.Database, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var dataType, isNullable, columnKey, extra, defaultValue sql.NullString

		if err := rows.Scan(
			&col.Name,
			&dataType,
			&col.Comment,
			&isNullable,
			&columnKey,
			&extra,
			&defaultValue,
		); err != nil {
			return nil, err
		}

		col.Type = dataType.String
		col.IsNullable = isNullable.String == "YES"
		col.IsPrimaryKey = columnKey.String == "PRI"
		col.IsAutoIncr = strings.Contains(extra.String, "auto_increment")
		col.DefaultValue = defaultValue.String

		// 转换为 Go 类型
		col.GoType = g.convertToGoType(col.Type, col.IsNullable)
		col.GoTag = g.generateGoTag(col)

		columns = append(columns, col)
	}

	return columns, nil
}

// getPrimaryKeys 获取主键信息
func (g *Generator) getPrimaryKeys(tableName string) ([]string, error) {
	query := `
		SELECT 
			COLUMN_NAME
		FROM 
			INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
		WHERE 
			TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY 
			ORDINAL_POSITION
	`

	rows, err := g.db.Query(query, g.config.Database, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, columnName)
	}

	return primaryKeys, nil
}

// convertToGoType 转换为 Go 类型
func (g *Generator) convertToGoType(dbType string, isNullable bool) string {
	dbType = strings.ToLower(dbType)

	var goType string
	switch {
	case strings.Contains(dbType, "int"):
		if strings.Contains(dbType, "bigint") {
			goType = "int64"
		} else {
			goType = "int"
		}
	case strings.Contains(dbType, "decimal"), strings.Contains(dbType, "numeric"):
		goType = "float64"
	case strings.Contains(dbType, "float"), strings.Contains(dbType, "double"):
		goType = "float64"
	case strings.Contains(dbType, "datetime"), strings.Contains(dbType, "timestamp"):
		goType = "time.Time"
	case strings.Contains(dbType, "date"):
		goType = "time.Time"
	case strings.Contains(dbType, "text"), strings.Contains(dbType, "varchar"), strings.Contains(dbType, "char"):
		goType = "string"
	case strings.Contains(dbType, "blob"):
		goType = "[]byte"
	case strings.Contains(dbType, "bool"):
		goType = "bool"
	default:
		goType = "string"
	}

	// 如果是可空字段，使用指针类型
	if isNullable && goType != "string" && goType != "[]byte" {
		goType = "*" + goType
	}

	return goType
}

// generateGoTag 生成 Go 标签
func (g *Generator) generateGoTag(col ColumnInfo) string {
	var tags []string

	// GORM 标签
	var gormTags []string
	gormTags = append(gormTags, fmt.Sprintf("column:%s", col.Name))

	if col.IsPrimaryKey {
		gormTags = append(gormTags, "primarykey")
	}

	if col.IsAutoIncr {
		gormTags = append(gormTags, "autoIncrement")
	}

	if !col.IsNullable {
		gormTags = append(gormTags, "not null")
	}

	// 根据类型添加其他标签
	if strings.Contains(col.Type, "varchar") || strings.Contains(col.Type, "char") {
		if size := g.extractSize(col.Type); size > 0 {
			gormTags = append(gormTags, fmt.Sprintf("size:%d", size))
		}
	}

	if col.Comment != "" {
		gormTags = append(gormTags, fmt.Sprintf("comment:%s", col.Comment))
	}

	tags = append(tags, fmt.Sprintf("gorm:\"%s\"", strings.Join(gormTags, ";")))

	// JSON 标签
	jsonName := g.toSnakeCase(col.Name)
	tags = append(tags, fmt.Sprintf("json:\"%s\"", jsonName))

	return strings.Join(tags, " ")
}

// extractSize 提取字段大小
func (g *Generator) extractSize(dbType string) int {
	re := regexp.MustCompile(`\((\d+)\)`)
	matches := re.FindStringSubmatch(dbType)
	if len(matches) > 1 {
		var size int
		fmt.Sscanf(matches[1], "%d", &size)
		return size
	}
	return 0
}

// toSnakeCase 转换为蛇形命名
func (g *Generator) toSnakeCase(s string) string {
	var result string
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return strings.ToLower(result)
}

// generateBaseModel 生成基础模型
func (g *Generator) generateBaseModel() error {
	tmpl := `package {{.Package}}

import (
	"time"
	"gorm.io/gorm"
)

// BaseModel 基础模型，包含所有模型的公共字段
type BaseModel struct {
	ID        uint           ` + "`gorm:\"primarykey\" json:\"id\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\" json:\"-\"`" + `
}
`

	t, err := template.New("base").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(g.config.Output, "base.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, map[string]string{
		"Package": g.config.Package,
	})
}

// generateTableModel 生成表模型
func (g *Generator) generateTableModel(table TableInfo) error {
	tmpl := `package {{.Package}}

import (
	"time"
)

// {{.StructName}} {{.Comment}}
type {{.StructName}} struct {
	{{- if .UseBaseModel}}
	BaseModel
	{{- end}}
	{{- range .Columns}}
	{{.GoName}} {{.GoType}} ` + "`{{.GoTag}}`" + `{{- if .Comment}} // {{.Comment}}{{end}}
	{{- end}}
}

// TableName 指定表名
func ({{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`

	t, err := template.New("table").Parse(tmpl)
	if err != nil {
		return err
	}

	// 准备模板数据
	data := map[string]interface{}{
		"Package":      g.config.Package,
		"StructName":   g.toCamelCase(table.Name),
		"TableName":    table.Name,
		"Comment":      table.Comment,
		"UseBaseModel": true,
		"Columns":      g.prepareColumns(table.Columns),
	}

	// 生成文件名
	fileName := g.toSnakeCase(table.Name) + ".go"
	filePath := filepath.Join(g.config.Output, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// prepareColumns 准备列数据
func (g *Generator) prepareColumns(columns []ColumnInfo) []map[string]interface{} {
	var result []map[string]interface{}
	for _, col := range columns {
		result = append(result, map[string]interface{}{
			"GoName": g.toCamelCase(col.Name),
			"GoType": col.GoType,
			"GoTag":  col.GoTag,
			"Comment": col.Comment,
		})
	}
	return result
}

// toCamelCase 转换为驼峰命名
func (g *Generator) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}
