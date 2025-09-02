# Go GORM 代码生成器

一个可复用的 MySQL -> GORM 代码生成器，支持生成 Model、Service、Router（Gin）代码，并允许自定义生成代码中各层的导入路径，方便在任何项目中落地。

## 安装

### 方式一：通过 go get 安装（推荐）

#### 安装最新版本
```bash
go install github.com/you/generator@latest
```

#### 安装特定版本
```bash
go install github.com/you/generator@v1.0.0
```

```bash
go install github.com/you/generator@latest
```

安装后，你可以在任何地方使用 `generator` 命令：

```bash
generator -database your_db -user root -password 123456 \
  -output internal/models -package models \
  -router -service \
  -router-output internal/router -service-output internal/services \
  -model-import github.com/you/yourapp/internal/models \
  -service-import github.com/you/yourapp/internal/services \
  -storage-import github.com/you/yourapp/internal/storage
```

### 方式二：作为可执行命令使用

```bash
# 运行最新版本
go run github.com/you/generator@latest -database your_db -user root -password 123456 \
  -output internal/models -package models \
  -router -service \
  -router-output internal/router -service-output internal/services \
  -model-import github.com/you/yourapp/internal/models \
  -service-import github.com/you/yourapp/internal/services \
  -storage-import github.com/you/yourapp/internal/storage

# 运行特定版本
go run github.com/you/generator@v1.0.0 -database your_db -user root -password 123456 \
  -output internal/models -package models \
  -router -service \
  -router-output internal/router -service-output internal/services \
  -model-import github.com/you/yourapp/internal/models \
  -service-import github.com/you/yourapp/internal/services \
  -storage-import github.com/you/yourapp/internal/storage
```

- 作为库在你的 Go 代码中调用：
```go
import (
    gen "github.com/you/generator/config"
)

cfg := &gen.Config{
    Host: "localhost",
    Port: 3306,
    User: "root",
    Password: "123456",
    Database: "your_db",
    Output: "internal/models",
    Package: "models",
    Tables: "",
    GenerateRouter: true,
    GenerateService: true,
    RouterOutput: "internal/router",
    ServiceOutput: "internal/services",
    ModelImportPath:   "github.com/you/yourapp/internal/models",
    ServiceImportPath: "github.com/you/yourapp/internal/services",
    StorageImportPath: "github.com/you/yourapp/internal/storage",
}

if err := gen.NewGenerator(cfg).Generate(); err != nil {
    panic(err)
}
```

## 配置文件

参考 `config.yaml`，也可通过 `-config` 指定：
```yaml
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "123456"
  database: "test_db"

output:
  path: "internal/models"
  package: "models"

options:
  generate_router: true
  generate_service: true

router:
  output: "internal/router"

service:
  output: "internal/services"

imports:
  model: "github.com/you/yourapp/internal/models"
  service: "github.com/you/yourapp/internal/services"
  storage: "github.com/you/yourapp/internal/storage"
```

命令行参数为空时，将回退到配置文件对应的字段。

## 主要命令行参数

- `-database` 数据库名（必填）
- `-host`/`-port`/`-user`/`-password` 数据库连接信息
- `-tables` 指定表（逗号分隔），为空生成全部
- `-output` 模型输出目录，`-package` 模型包名
- `-router` 是否生成 Router，`-router-output` Router 输出目录
- `-service` 是否生成 Service，`-service-output` Service 输出目录
- `-model-import` 生成代码中 model 包的导入路径
- `-service-import` 生成代码中 service 包的导入路径
- `-storage-import` 生成代码中 storage 根包的导入路径
- `-config` 配置文件路径（默认 `config.yaml`）

## 生成内容说明

- Model：包含基础 `BaseModel` 与每张表的结构体定义、`TableName()`
- Service：CRUD、分页、可选搜索/唯一字段方法，依赖 `storage/mysql` 的 `DB`
- Router：Gin handler，包含增删改查、可选搜索、分页封装

## 嵌入方式建议

- 你的项目需要准备：
  - `internal/storage/mysql` 包并暴露 `DB *gorm.DB`
  - `internal/models`、`internal/services`、`internal/router` 目录（可根据需要调整）
- 在运行生成器时，传入上述目录对应的导入路径，生成代码会直接可用。

## 版本信息

当前版本：`v1.0.0`

查看 [CHANGELOG.md](CHANGELOG.md) 了解详细版本变更。

## 许可证

MIT
