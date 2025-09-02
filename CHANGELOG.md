# 更新日志

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范。

## [未发布]

### 新增
- 初始版本

## [v1.0.0] - 2024-09-02

### 新增
- 🎉 初始版本发布
- ✨ 支持 MySQL 数据库表结构解析
- ✨ 自动生成 GORM Model 代码
- ✨ 支持生成 Service 层代码
- ✨ 支持生成 Router 层代码（Gin）
- ✨ 支持自定义导入路径配置
- ✨ 支持配置文件方式配置
- ✨ 支持命令行参数配置
- ✨ 支持指定表名生成
- ✨ 支持软删除功能
- ✨ 支持 JSON 和 GORM 标签生成

### 特性
- 🔧 灵活的配置系统（命令行 + 配置文件）
- 📝 详细的帮助信息和示例
- 🎯 支持作为库或可执行文件使用
- 📦 支持通过 go get 安装

## 版本说明

- **主版本号**：不兼容的 API 修改
- **次版本号**：向下兼容的功能性新增
- **修订号**：向下兼容的问题修正

## 安装特定版本

```bash
# 安装最新版本
go install github.com/you/generator@latest

# 安装特定版本
go install github.com/you/generator@v1.0.0

# 安装特定版本（go get 方式）
go get github.com/you/generator@v1.0.0
```
