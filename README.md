# feishu-docs

基于飞书开放平台 Golang SDK 的命令行工具，支持文档(Docx)和知识库(Wiki)的增删改查操作。

## 安装

### 通过 Homebrew（推荐）

```bash
brew tap KQAR/tap
brew install feishu-docs
```

升级：

```bash
brew upgrade feishu-docs
```

### 从源码编译

```bash
git clone https://github.com/KQAR/feishu-docs.git
cd feishu-docs
go install .
```

## 配置

运行初始化命令创建配置模板：

```bash
feishu-docs init
```

编辑 `~/config/feishu-docs/config.json`，填入飞书应用凭证：

```json
{
  "app_id": "cli_xxxx",
  "app_secret": "xxxx"
}
```

> 凭证获取方式：前往 [飞书开放平台开发者控制台](https://open.feishu.cn/app) 创建应用并获取。

## 使用

### 文档操作

```bash
# 创建文档
feishu-docs doc create --title "我的文档" --folder-token <folder_token>

# 获取文档信息
feishu-docs doc get --id <document_id>

# 获取文档纯文本内容
feishu-docs doc content --id <document_id>

# 列出文档所有块
feishu-docs doc blocks --id <document_id>

# 获取单个块的详细信息
feishu-docs doc block --doc-id <document_id> --block-id <block_id>

# 插入内容块 (支持 text/heading1~9/bullet/ordered/code/todo)
feishu-docs doc insert --doc-id <document_id> --text "Hello World"
feishu-docs doc insert --doc-id <document_id> --text "标题" --type heading2 --index 0

# 更新块内容
feishu-docs doc update --doc-id <document_id> --block-id <block_id> --text "新内容"

# 删除文档子块
feishu-docs doc delete-blocks --doc-id <document_id> --block-id <block_id> --start 0 --end 2
```

### 知识库(Wiki)操作

```bash
# 列出知识空间
feishu-docs wiki spaces

# 获取知识空间信息
feishu-docs wiki space --id <space_id>

# 获取节点信息
feishu-docs wiki node --token <node_token>

# 列出子节点
feishu-docs wiki nodes --space-id <space_id> [--parent <parent_node_token>]

# 创建节点
feishu-docs wiki create --space-id <space_id> --title "新页面" [--obj-type docx] [--parent <parent_token>]

# 移动节点
feishu-docs wiki move --space-id <space_id> --node-token <token> --target-parent <target_token>

# 复制节点
feishu-docs wiki copy --space-id <space_id> --node-token <token> --target-parent <target_token> --target-space <target_space_id>
```

## 权限要求

请确保飞书应用已开通以下权限：

- `docx:document` - 文档读写权限
- `wiki:wiki` - 知识库读写权限

## 发布新版本

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

推送 tag 后，GitHub Actions 会自动构建多平台二进制并更新 Homebrew tap。

## 依赖

- [oapi-sdk-go v3](https://github.com/larksuite/oapi-sdk-go) - 飞书开放平台 Golang SDK
- [cobra](https://github.com/spf13/cobra) - CLI 框架
