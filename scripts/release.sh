#!/bin/bash

# 版本发布脚本
# 使用方法: ./scripts/release.sh <version>

set -e

if [ $# -eq 0 ]; then
    echo "使用方法: $0 <version>"
    echo "示例: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# 验证版本格式
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "错误: 版本格式不正确，应该是 vX.Y.Z 格式"
    echo "示例: v1.0.0, v2.1.3"
    exit 1
fi

echo "准备发布版本: $VERSION"

# 检查是否有未提交的更改
if [ -n "$(git status --porcelain)" ]; then
    echo "错误: 有未提交的更改，请先提交所有更改"
    git status --short
    exit 1
fi

# 检查当前分支
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    echo "警告: 当前不在 main/master 分支，当前分支: $CURRENT_BRANCH"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 检查版本是否已存在
if git tag -l | grep -q "^$VERSION$"; then
    echo "错误: 版本 $VERSION 已经存在"
    exit 1
fi

# 更新 go.mod 中的版本信息（如果有的话）
# 这里可以添加更新 go.mod 的逻辑

# 创建版本标签
echo "创建版本标签: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"

# 推送标签到远程仓库
echo "推送标签到远程仓库..."
git push origin "$VERSION"

echo "✅ 版本 $VERSION 发布成功！"
echo ""
echo "用户现在可以通过以下方式安装:"
echo "  go install github.com/you/generator@$VERSION"
echo "  go get github.com/you/generator@$VERSION"
echo ""
echo "或者安装最新版本:"
echo "  go install github.com/you/generator@latest"
