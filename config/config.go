package generator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ConfigFile 配置文件结构
type ConfigFile struct {
	Database DatabaseConfig `yaml:"database"`
	Output   OutputConfig   `yaml:"output"`
	Tables   string         `yaml:"tables,omitempty"`
	Options  OptionsConfig  `yaml:"options"`
	Router   RouterConfig   `yaml:"router"`
	Service  ServiceConfig  `yaml:"service"`
	Imports  ImportConfig   `yaml:"imports"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Path    string `yaml:"path"`
	Package string `yaml:"package"`
}

// OptionsConfig 生成选项
type OptionsConfig struct {
	GenerateBaseModel   bool `yaml:"generate_base_model"`
	UseSoftDelete       bool `yaml:"use_soft_delete"`
	GenerateJSONTags    bool `yaml:"generate_json_tags"`
	GenerateGORMTags    bool `yaml:"generate_gorm_tags"`
	GenerateComments    bool `yaml:"generate_comments"`
	GenerateRouter      bool `yaml:"generate_router"`
	GenerateService     bool `yaml:"generate_service"`
}

// ImportConfig 导入路径配置
type ImportConfig struct {
	Model   string `yaml:"model"`
	Service string `yaml:"service"`
	Storage string `yaml:"storage"`
}

// RouterConfig Router配置
type RouterConfig struct {
	Output string `yaml:"output"`
}

// ServiceConfig Service配置
type ServiceConfig struct {
	Output string `yaml:"output"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*ConfigFile, error) {
	if configPath == "" {
		configPath = "cmd/generator/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// MergeConfig 合并命令行参数和配置文件
func MergeConfig(cmdConfig *Config, fileConfig *ConfigFile) *Config {
	result := &Config{
		Host:            cmdConfig.Host,
		Port:            cmdConfig.Port,
		User:            cmdConfig.User,
		Password:        cmdConfig.Password,
		Database:        cmdConfig.Database,
		Output:          cmdConfig.Output,
		Package:         cmdConfig.Package,
		Tables:          cmdConfig.Tables,
		GenerateRouter:  cmdConfig.GenerateRouter,
		GenerateService: cmdConfig.GenerateService,
		RouterOutput:    cmdConfig.RouterOutput,
		ServiceOutput:   cmdConfig.ServiceOutput,
		ModelImportPath:   cmdConfig.ModelImportPath,
		ServiceImportPath: cmdConfig.ServiceImportPath,
		StorageImportPath: cmdConfig.StorageImportPath,
	}

	// 如果命令行参数为空，使用配置文件的值
	if result.Host == "" {
		result.Host = fileConfig.Database.Host
	}
	if result.Port == 0 {
		result.Port = fileConfig.Database.Port
	}
	if result.User == "" {
		result.User = fileConfig.Database.User
	}
	if result.Password == "" {
		result.Password = fileConfig.Database.Password
	}
	if result.Database == "" {
		result.Database = fileConfig.Database.Database
	}
	if result.Output == "" {
		result.Output = fileConfig.Output.Path
	}
	if result.Package == "" {
		result.Package = fileConfig.Output.Package
	}
	if result.Tables == "" {
		result.Tables = fileConfig.Tables
	}
	if !result.GenerateRouter {
		result.GenerateRouter = fileConfig.Options.GenerateRouter
	}
	if !result.GenerateService {
		result.GenerateService = fileConfig.Options.GenerateService
	}
	if result.RouterOutput == "" {
		result.RouterOutput = fileConfig.Router.Output
	}
	if result.ServiceOutput == "" {
		result.ServiceOutput = fileConfig.Service.Output
	}
	// 导入路径合并
	if result.ModelImportPath == "" {
		result.ModelImportPath = fileConfig.Imports.Model
	}
	if result.ServiceImportPath == "" {
		result.ServiceImportPath = fileConfig.Imports.Service
	}
	if result.StorageImportPath == "" {
		result.StorageImportPath = fileConfig.Imports.Storage
	}

	return result
}
