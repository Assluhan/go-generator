package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// generateRouters 生成 Router 代码
func (g *Generator) generateRouters(tables []TableInfo) error {
	// 创建 Router 输出目录
	if err := os.MkdirAll(g.config.RouterOutput, 0755); err != nil {
		return fmt.Errorf("创建 Router 输出目录失败: %w", err)
	}

	// 生成 Router 基础文件
	if err := g.generateRouterBase(); err != nil {
		return fmt.Errorf("生成 Router 基础文件失败: %w", err)
	}

	// 为每个表生成 Router
	for _, table := range tables {
		if err := g.generateTableRouter(table); err != nil {
			log.Printf("生成表 %s 的 Router 失败: %v", table.Name, err)
			continue
		}
		fmt.Printf("生成表 %s 的 Router 成功\n", table.Name)
	}

	return nil
}

// generateRouterBase 生成 Router 基础文件
func (g *Generator) generateRouterBase() error {
	tmpl := `package router

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         ` + "`json:\"code\"`" + `
	Message string      ` + "`json:\"message\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// GetPageParams 获取分页参数
func GetPageParams(c *gin.Context) (int, int) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return page, pageSize
}

// GetIDParam 获取ID参数
func GetIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
`

	file, err := os.Create(filepath.Join(g.config.RouterOutput, "base.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(tmpl)
	return err
}

// generateTableRouter 生成表 Router
func (g *Generator) generateTableRouter(table TableInfo) error {
	tmpl := `package router

import (
	"net/http"
	"strconv"

	"{{.ModelPackage}}"
	"{{.ServicePackage}}"
	"github.com/gin-gonic/gin"
)

// {{.HandlerName}} {{.Comment}}处理器
type {{.HandlerName}} struct {
	{{.ServiceVarName}} *{{.ServiceName}}
}

// New{{.HandlerName}} 创建{{.Comment}}处理器
func New{{.HandlerName}}() *{{.HandlerName}} {
	return &{{.HandlerName}}{
		{{.ServiceVarName}}: {{.ServicePackage}}.New{{.ServiceName}}(),
	}
}

// Create{{.ModelName}} 创建{{.Comment}}
func (h *{{.HandlerName}}) Create{{.ModelName}}(c *gin.Context) {
	var {{.ModelVarName}} {{.ModelName}}
	if err := c.ShouldBindJSON(&{{.ModelVarName}}); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	{{- if .HasUniqueFields}}
	{{- range .UniqueFields}}
	// 检查{{.Comment}}是否已存在
	if {{.ModelVarName}}.{{.GoName}} != "" {
		if _, err := h.{{$.ServiceVarName}}.GetBy{{.GoName}}({{.ModelVarName}}.{{.GoName}}); err == nil {
			Error(c, 409, "{{.Comment}}已存在")
			return
		}
	}
	{{- end}}
	{{- end}}

	if err := h.{{.ServiceVarName}}.Create(&{{.ModelVarName}}); err != nil {
		Error(c, 500, "创建{{.Comment}}失败: "+err.Error())
		return
	}

	Success(c, {{.ModelVarName}})
}

// Get{{.ModelName}} 获取{{.Comment}}
func (h *{{.HandlerName}}) Get{{.ModelName}}(c *gin.Context) {
	id, err := GetIDParam(c)
	if err != nil {
		Error(c, 400, "无效的ID")
		return
	}

	{{.ModelVarName}}, err := h.{{.ServiceVarName}}.GetByID(id)
	if err != nil {
		Error(c, 404, "{{.Comment}}不存在")
		return
	}

	Success(c, {{.ModelVarName}})
}

// Update{{.ModelName}} 更新{{.Comment}}
func (h *{{.HandlerName}}) Update{{.ModelName}}(c *gin.Context) {
	id, err := GetIDParam(c)
	if err != nil {
		Error(c, 400, "无效的ID")
		return
	}

	var updateData {{.ModelName}}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		Error(c, 400, "请求参数错误: "+err.Error())
		return
	}

	{{.ModelVarName}}, err := h.{{.ServiceVarName}}.GetByID(id)
	if err != nil {
		Error(c, 404, "{{.Comment}}不存在")
		return
	}

	// 更新字段
	{{- range .UpdateableFields}}
	if updateData.{{.GoName}} != {{.ZeroValue}} {
		{{$.ModelVarName}}.{{.GoName}} = updateData.{{.GoName}}
	}
	{{- end}}

	if err := h.{{.ServiceVarName}}.Update({{.ModelVarName}}); err != nil {
		Error(c, 500, "更新{{.Comment}}失败: "+err.Error())
		return
	}

	Success(c, {{.ModelVarName}})
}

// Delete{{.ModelName}} 删除{{.Comment}}
func (h *{{.HandlerName}}) Delete{{.ModelName}}(c *gin.Context) {
	id, err := GetIDParam(c)
	if err != nil {
		Error(c, 400, "无效的ID")
		return
	}

	if err := h.{{.ServiceVarName}}.Delete(id); err != nil {
		Error(c, 500, "删除{{.Comment}}失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}

// List{{.ModelName}}s 获取{{.Comment}}列表
func (h *{{.HandlerName}}) List{{.ModelName}}s(c *gin.Context) {
	page, pageSize := GetPageParams(c)

	{{.ModelVarName}}s, total, err := h.{{.ServiceVarName}}.List(page, pageSize)
	if err != nil {
		Error(c, 500, "获取{{.Comment}}列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"list":      {{.ModelVarName}}s,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

{{- if .HasSearchFields}}
// Search{{.ModelName}}s 搜索{{.Comment}}
func (h *{{.HandlerName}}) Search{{.ModelName}}s(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		Error(c, 400, "搜索关键词不能为空")
		return
	}

	page, pageSize := GetPageParams(c)

	{{.ModelVarName}}s, total, err := h.{{.ServiceVarName}}.Search(keyword, page, pageSize)
	if err != nil {
		Error(c, 500, "搜索{{.Comment}}失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"list":      {{.ModelVarName}}s,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
{{- end}}

// Register{{.ModelName}}Routes 注册{{.Comment}}路由
func Register{{.ModelName}}Routes(r *gin.RouterGroup) {
	handler := New{{.HandlerName}}()
	
	{{.RouteGroup}} := r.Group("/{{.RoutePath}}")
	{
		{{.RouteGroup}}.POST("", handler.Create{{.ModelName}})
		{{.RouteGroup}}.GET("", handler.List{{.ModelName}}s)
		{{.RouteGroup}}.GET("/:id", handler.Get{{.ModelName}})
		{{.RouteGroup}}.PUT("/:id", handler.Update{{.ModelName}})
		{{.RouteGroup}}.DELETE("/:id", handler.Delete{{.ModelName}})
		{{- if .HasSearchFields}}
		{{.RouteGroup}}.GET("/search", handler.Search{{.ModelName}}s)
		{{- end}}
	}
}
`

	t, err := template.New("router").Parse(tmpl)
	if err != nil {
		return err
	}

	// 准备模板数据
	data := map[string]interface{}{
		"ModelPackage":     g.config.ModelImportPath,
		"ServicePackage":   g.config.ServiceImportPath,
		"HandlerName":      g.toCamelCase(table.Name) + "Handler",
		"ServiceName":      g.toCamelCase(table.Name) + "Service",
		"ServiceVarName":   g.toLowerCamelCase(table.Name) + "Service",
		"ModelName":        g.toCamelCase(table.Name),
		"ModelVarName":     g.toLowerCamelCase(table.Name),
		"Comment":          table.Comment,
		"RouteGroup":       g.toLowerCamelCase(table.Name) + "Group",
		"RoutePath":        g.toSnakeCase(table.Name),
		"UniqueFields":     g.getUniqueFields(table.Columns),
		"UpdateableFields": g.getUpdateableFields(table.Columns),
		"SearchFields":     g.getSearchFields(table.Columns),
		"HasUniqueFields":  len(g.getUniqueFields(table.Columns)) > 0,
		"HasSearchFields":  len(g.getSearchFields(table.Columns)) > 0,
	}

	// 生成文件名
	fileName := g.toSnakeCase(table.Name) + "_router.go"
	filePath := filepath.Join(g.config.RouterOutput, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// getUpdateableFields 获取可更新字段
func (g *Generator) getUpdateableFields(columns []ColumnInfo) []map[string]interface{} {
	var result []map[string]interface{}
	for _, col := range columns {
		// 排除主键、创建时间等不可更新字段
		if !col.IsPrimaryKey && 
		   !strings.Contains(strings.ToLower(col.Name), "created_at") &&
		   !strings.Contains(strings.ToLower(col.Name), "id") {
			result = append(result, map[string]interface{}{
				"GoName":     g.toCamelCase(col.Name),
				"ZeroValue":  g.getZeroValue(col.GoType),
			})
		}
	}
	return result
}

// getZeroValue 获取零值
func (g *Generator) getZeroValue(goType string) string {
	switch goType {
	case "string":
		return `""`
	case "int", "int64":
		return "0"
	case "float64":
		return "0.0"
	case "bool":
		return "false"
	case "time.Time":
		return "time.Time{}"
	default:
		if strings.HasPrefix(goType, "*") {
			return "nil"
		}
		return "nil"
	}
}
