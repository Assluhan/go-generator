package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go-web/cmd/generator/generator"
)

func main() {
	// 命令行参数
	var (
		host           = flag.String("host", "", "数据库主机")
		port           = flag.Int("port", 0, "数据库端口")
		user           = flag.String("user", "", "数据库用户名")
		password       = flag.String("password", "", "数据库密码")
		database       = flag.String("database", "", "数据库名")
		output         = flag.String("output", "", "输出目录")
		tables         = flag.String("tables", "", "指定表名，多个表用逗号分隔，为空则生成所有表")
		package        = flag.String("package", "", "生成的包名")
		configFile     = flag.String("config", "config.yaml", "配置文件路径")
		generateRouter = flag.Bool("router", false, "是否生成Router代码")
		generateService = flag.Bool("service", false, "是否生成Service代码")
		routerOutput   = flag.String("router-output", "", "Router输出目录")
		serviceOutput  = flag.String("service-output", "", "Service输出目录")
		modelImport    = flag.String("model-import", "", "模型包导入路径，例如: github.com/your/app/internal/models")
		serviceImport  = flag.String("service-import", "", "服务包导入路径，例如: github.com/your/app/internal/services")
		storageImport  = flag.String("storage-import", "", "存储包导入路径根，例如: github.com/your/app/internal/storage")
		help           = flag.Bool("help", false, "显示帮助信息")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// 加载配置文件
	fileConfig, err := generator.LoadConfig(*configFile)
	if err != nil {
		log.Printf("警告: 加载配置文件失败: %v", err)
		log.Println("将使用命令行参数")
	}

	// 创建命令行配置
	cmdConfig := &generator.Config{
		Host:            *host,
		Port:            *port,
		User:            *user,
		Password:        *password,
		Database:        *database,
		Output:          *output,
		Package:         *package,
		Tables:          *tables,
		GenerateRouter:  *generateRouter,
		GenerateService: *generateService,
		RouterOutput:    *routerOutput,
		ServiceOutput:   *serviceOutput,
		ModelImportPath:   *modelImport,
		ServiceImportPath: *serviceImport,
		StorageImportPath: *storageImport,
	}

	// 合并配置
	var finalConfig *generator.Config
	if fileConfig != nil {
		finalConfig = generator.MergeConfig(cmdConfig, fileConfig)
	} else {
		finalConfig = cmdConfig
	}

	// 验证必需参数
	if finalConfig.Database == "" {
		log.Fatal("请指定数据库名 (-database 或配置文件)")
	}

	// 创建生成器
	gen := generator.NewGenerator(finalConfig)

	// 生成代码
	if err := gen.Generate(); err != nil {
		log.Fatalf("生成代码失败: %v", err)
	}

	fmt.Println("代码生成完成！")
}

func showHelp() {
	fmt.Println("GORM 代码生成器")
	fmt.Println("用法: go run cmd/generator/main.go [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -host string")
	fmt.Println("        数据库主机 (默认: localhost)")
	fmt.Println("  -port int")
	fmt.Println("        数据库端口 (默认: 3306)")
	fmt.Println("  -user string")
	fmt.Println("        数据库用户名 (默认: root)")
	fmt.Println("  -password string")
	fmt.Println("        数据库密码")
	fmt.Println("  -database string")
	fmt.Println("        数据库名 (必需)")
	fmt.Println("  -output string")
	fmt.Println("        输出目录 (默认: internal/models)")
	fmt.Println("  -tables string")
	fmt.Println("        指定表名，多个表用逗号分隔，为空则生成所有表")
	fmt.Println("  -package string")
	fmt.Println("        生成的包名 (默认: models)")
	fmt.Println("  -router")
	fmt.Println("        是否生成Router代码")
	fmt.Println("  -service")
	fmt.Println("        是否生成Service代码")
	fmt.Println("  -router-output string")
	fmt.Println("        Router输出目录")
	fmt.Println("  -service-output string")
	fmt.Println("        Service输出目录")
	fmt.Println("  -model-import string")
	fmt.Println("        模型包导入路径，例如: github.com/your/app/internal/models")
	fmt.Println("  -service-import string")
	fmt.Println("        服务包导入路径，例如: github.com/your/app/internal/services")
	fmt.Println("  -storage-import string")
	fmt.Println("        存储包导入路径根，例如: github.com/your/app/internal/storage")
	fmt.Println("  -config string")
	fmt.Println("        配置文件路径")
	fmt.Println("  -help")
	fmt.Println("        显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run cmd/generator/main.go -database test_db -password 123456")
	fmt.Println("  go run cmd/generator/main.go -database test_db -tables users,articles -output ./models")
	fmt.Println("  go run cmd/generator/main.go -database test_db -password 123456 -router -service")
	fmt.Println("  go run cmd/generator/main.go -database test_db -password 123456 -router -service -tables users,articles")
}
