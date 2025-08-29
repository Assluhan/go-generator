package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// generateServices 生成 Service 代码
func (g *Generator) generateServices(tables []TableInfo) error {
	// 创建 Service 输出目录
	if err := os.MkdirAll(g.config.ServiceOutput, 0755); err != nil {
		return fmt.Errorf("创建 Service 输出目录失败: %w", err)
	}

	// 生成 Service 基础文件
	if err := g.generateServiceBase(); err != nil {
		return fmt.Errorf("生成 Service 基础文件失败: %w", err)
	}

	// 为每个表生成 Service
	for _, table := range tables {
		if err := g.generateTableService(table); err != nil {
			log.Printf("生成表 %s 的 Service 失败: %v", table.Name, err)
			continue
		}
		fmt.Printf("生成表 %s 的 Service 成功\n", table.Name)
	}

	return nil
}

// generateServiceBase 生成 Service 基础文件
func (g *Generator) generateServiceBase() error {
	tmpl := `package services

import (
	"errors"
	"gorm.io/gorm"
)

// BaseService 基础服务接口
type BaseService interface {
	Create(model interface{}) error
	GetByID(id uint) (interface{}, error)
	Update(model interface{}) error
	Delete(id uint) error
	List(page, pageSize int) ([]interface{}, int64, error)
}

// ServiceError 服务错误
type ServiceError struct {
	Code    int
	Message string
}

func (e ServiceError) Error() string {
	return e.Message
}

// NewServiceError 创建服务错误
func NewServiceError(code int, message string) error {
	return ServiceError{
		Code:    code,
		Message: message,
	}
}

// IsNotFound 检查是否为未找到错误
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
`

	file, err := os.Create(filepath.Join(g.config.ServiceOutput, "base.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(tmpl)
	return err
}

// generateTableService 生成表 Service
func (g *Generator) generateTableService(table TableInfo) error {
	tmpl := `package services

import (
	"{{.ModelPackage}}"
	mysqlx "{{.StoragePackage}}/mysql"
	"gorm.io/gorm"
)

// {{.ServiceName}} {{.Comment}}服务
type {{.ServiceName}} struct{}

// New{{.ServiceName}} 创建{{.Comment}}服务实例
func New{{.ServiceName}}() *{{.ServiceName}} {
	return &{{.ServiceName}}{}
}

// Create 创建{{.Comment}}
func (s *{{.ServiceName}}) Create({{.ModelVarName}} *{{.ModelName}}) error {
	return mysqlx.DB.Create({{.ModelVarName}}).Error
}

// GetByID 根据ID获取{{.Comment}}
func (s *{{.ServiceName}}) GetByID(id uint) (*{{.ModelName}}, error) {
	var {{.ModelVarName}} {{.ModelName}}
	err := mysqlx.DB.First(&{{.ModelVarName}}, id).Error
	if err != nil {
		return nil, err
	}
	return &{{.ModelVarName}}, nil
}

{{- if .HasUniqueFields}}
{{- range .UniqueFields}}
// GetBy{{.GoName}} 根据{{.Comment}}获取{{$.Comment}}
func (s *{{$.ServiceName}}) GetBy{{.GoName}}({{.VarName}} {{.GoType}}) (*{{$.ModelName}}, error) {
	var {{$.ModelVarName}} {{$.ModelName}}
	err := mysqlx.DB.Where("{{.DBName}} = ?", {{.VarName}}).First(&{{$.ModelVarName}}).Error
	if err != nil {
		return nil, err
	}
	return &{{$.ModelVarName}}, nil
}
{{- end}}
{{- end}}

// Update 更新{{.Comment}}
func (s *{{.ServiceName}}) Update({{.ModelVarName}} *{{.ModelName}}) error {
	return mysqlx.DB.Save({{.ModelVarName}}).Error
}

// Delete 删除{{.Comment}}
func (s *{{.ServiceName}}) Delete(id uint) error {
	return mysqlx.DB.Delete(&{{.ModelName}}{}, id).Error
}

// List 获取{{.Comment}}列表
func (s *{{.ServiceName}}) List(page, pageSize int) ([]{{.ModelName}}, int64, error) {
	var {{.ModelVarName}}s []{{.ModelName}}
	var total int64

	// 获取总数
	err := mysqlx.DB.Model(&{{.ModelName}}{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err = mysqlx.DB.Offset(offset).Limit(pageSize).Find(&{{.ModelVarName}}s).Error
	if err != nil {
		return nil, 0, err
	}

	return {{.ModelVarName}}s, total, nil
}

{{- if .HasSearchFields}}
// Search 搜索{{.Comment}}
func (s *{{.ServiceName}}) Search(keyword string, page, pageSize int) ([]{{.ModelName}}, int64, error) {
	var {{.ModelVarName}}s []{{.ModelName}}
	var total int64

	query := mysqlx.DB.Model(&{{.ModelName}}{})
	{{- range .SearchFields}}
	query = query.Where("{{.DBName}} LIKE ?", "%"+keyword+"%")
	{{- end}}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Find(&{{.ModelVarName}}s).Error
	if err != nil {
		return nil, 0, err
	}

	return {{.ModelVarName}}s, total, nil
}
{{- end}}
`

	t, err := template.New("service").Parse(tmpl)
	if err != nil {
		return err
	}

	// 准备模板数据
	data := map[string]interface{}{
		"ModelPackage":    g.config.ModelImportPath,
		"StoragePackage":  g.config.StorageImportPath,
		"ServiceName":     g.toCamelCase(table.Name) + "Service",
		"ModelName":       g.toCamelCase(table.Name),
		"ModelVarName":    g.toLowerCamelCase(table.Name),
		"Comment":         table.Comment,
		"UniqueFields":    g.getUniqueFields(table.Columns),
		"SearchFields":    g.getSearchFields(table.Columns),
		"HasUniqueFields": len(g.getUniqueFields(table.Columns)) > 0,
		"HasSearchFields": len(g.getSearchFields(table.Columns)) > 0,
	}

	// 生成文件名
	fileName := g.toSnakeCase(table.Name) + "_service.go"
	filePath := filepath.Join(g.config.ServiceOutput, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// toLowerCamelCase 转换为小驼峰命名
func (g *Generator) toLowerCamelCase(s string) string {
	camel := g.toCamelCase(s)
	if len(camel) > 0 {
		return strings.ToLower(camel[:1]) + camel[1:]
	}
	return camel
}

// getUniqueFields 获取唯一字段
func (g *Generator) getUniqueFields(columns []ColumnInfo) []map[string]interface{} {
	var result []map[string]interface{}
	for _, col := range columns {
		// 检查是否为唯一字段（通过字段名或注释判断）
		if strings.Contains(strings.ToLower(col.Name), "username") ||
			strings.Contains(strings.ToLower(col.Name), "email") ||
			strings.Contains(strings.ToLower(col.Name), "phone") ||
			strings.Contains(strings.ToLower(col.Comment), "唯一") {
			result = append(result, map[string]interface{}{
				"GoName":  g.toCamelCase(col.Name),
				"GoType":  col.GoType,
				"VarName": g.toLowerCamelCase(col.Name),
				"DBName":  col.Name,
				"Comment": col.Comment,
			})
		}
	}
	return result
}

// getSearchFields 获取搜索字段
func (g *Generator) getSearchFields(columns []ColumnInfo) []map[string]interface{} {
	var result []map[string]interface{}
	for _, col := range columns {
		// 字符串类型字段可用于搜索
		if col.GoType == "string" || strings.Contains(col.GoType, "string") {
			result = append(result, map[string]interface{}{
				"DBName": col.Name,
			})
		}
	}
	return result
}
