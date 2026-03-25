# feishu-docs-cli

基于飞书开放平台 Golang SDK 的命令行工具，支持文档(Docx)和知识库(Wiki)的增删改查操作。

> 🚀 **最新更新**：Wiki 链接自动解析、增强 Markdown（代码块/分割线/引用）、stdin 管道输入

## 特性

- ✅ **文档 CRUD**：创建、读取、更新、删除文档
- ✅ **统一写操作入口**：通过 `doc update` 统一管理插入、更新、删除和表格编辑
- ✅ **Wiki 自动解析**：所有 `--doc-id` / `--id` 直接接受 wiki 链接，自动解析为文档 ID
- ✅ **Markdown 支持**：标题、列表、代码块、分割线、引用、待办
- ✅ **stdin 管道输入**：`--markdown -` 和 `--data -` 支持从管道读取内容
- ✅ **API 限流保护**：自动节流 + 429 指数退避重试，批量操作不再失败
- ✅ **知识库管理**：空间、节点的完整操作
- ✅ **跨平台**：支持 macOS、Linux (amd64/arm64)

## 安装

### 通过 Homebrew（推荐）

```bash
brew tap KQAR/tap
brew install feishu-docs-cli
```

升级：

```bash
brew upgrade feishu-docs-cli
```

### 从源码编译

```bash
git clone https://github.com/KQAR/feishu-docs-cli.git
cd feishu-docs-cli
go install .
```

## 配置

运行初始化命令创建配置模板：

```bash
feishu-docs-cli init
```

编辑 `~/.config/feishu-docs-cli/config.json`，填入飞书应用凭证：

```json
{
  "app_id": "cli_xxxx",
  "app_secret": "xxxx"
}
```

> 凭证获取方式：前往 [飞书开放平台开发者控制台](https://open.feishu.cn/app) 创建应用并获取。

## 使用

### 文档操作

所有 `--id` 和 `--doc-id` 参数都支持直接传入 wiki 链接，自动解析为文档 ID：

```bash
# 创建文档
feishu-docs-cli doc create --title "我的文档" --folder-token <folder_token>

# 获取文档信息（支持 wiki 链接）
feishu-docs-cli doc get --id <document_id>
feishu-docs-cli doc get --id "https://xxx.feishu.cn/wiki/TOKEN"

# 获取文档纯文本内容
feishu-docs-cli doc content --id <document_id>

# 列出文档所有块
feishu-docs-cli doc blocks --id <document_id>

# 获取单个块的详细信息
feishu-docs-cli doc block --doc-id <document_id> --block-id <block_id>
```

### 写操作（doc update）

```bash
# 追加 Markdown（支持代码块、分割线、引用、列表、待办）
feishu-docs-cli doc update append --doc-id <document_id> --markdown "# 标题

> 引用内容

---

\`\`\`
代码块
\`\`\`

- 无序列表
1. 有序列表
[x] 已完成待办"

# 从 stdin 管道读取 Markdown
cat content.md | feishu-docs-cli doc update append --doc-id <document_id> --markdown -

# 追加到指定父块
feishu-docs-cli doc update append --doc-id <document_id> --block-id <parent_block_id> --markdown "- 列表项"

# 插入单个文本块 (支持 text/heading1~9/bullet/ordered/code/todo)
feishu-docs-cli doc update insert --doc-id <document_id> --text "Hello World"
feishu-docs-cli doc update insert --doc-id <document_id> --text "标题" --type heading2 --index 0

# 按 block-id 更新文本块
feishu-docs-cli doc update set-text --doc-id <document_id> --block-id <block_id> --text "新内容"

# 删除某个父块下指定范围的子块
feishu-docs-cli doc update delete --doc-id <document_id> --block-id <block_id> --start 0 --end 2
```

### 表格操作（doc update table）

```bash
# 创建表格
feishu-docs-cli doc update table create --doc-id <document_id> --rows 2 --cols 3
feishu-docs-cli doc update table create --doc-id <document_id> --data $'姓名\t角色\n张三\tOwner'
feishu-docs-cli doc update table create --doc-id <document_id> --data '[["姓名","角色"],["张三","Owner"]]' --header-row

# 从 stdin 管道读取表格数据
echo -e "A\tB\n1\t2" | feishu-docs-cli doc update table create --doc-id <document_id> --data -

# 查看表格（默认 JSON，支持 --format tsv/table）
feishu-docs-cli doc update table show --doc-id <document_id> --table-id <table_block_id>
feishu-docs-cli doc update table show --doc-id <document_id> --table-id <table_block_id> --format tsv
feishu-docs-cli doc update table show --doc-id <document_id> --table-id <table_block_id> --format table

# 重写整张表格
feishu-docs-cli doc update table write --doc-id <document_id> --table-id <table_block_id> --data $'A\tB\n1\t2'

# 更新单元格
feishu-docs-cli doc update table set-cell --doc-id <document_id> --table-id <table_block_id> --row 1 --col 0 --text "李四"

# 插入/删除行列
feishu-docs-cli doc update table insert-row --doc-id <document_id> --table-id <table_block_id> --row-index -1 --data $'新值1\t新值2'
feishu-docs-cli doc update table insert-column --doc-id <document_id> --table-id <table_block_id> --column-index 1 --data $'表头\n内容'
feishu-docs-cli doc update table delete-rows --doc-id <document_id> --table-id <table_block_id> --start 1 --count 1
feishu-docs-cli doc update table delete-columns --doc-id <document_id> --table-id <table_block_id> --start 2 --count 1

# 合并/取消合并单元格
feishu-docs-cli doc update table merge --doc-id <document_id> --table-id <table_block_id> --row-start 0 --row-end 1 --column-start 0 --column-end 2
feishu-docs-cli doc update table unmerge --doc-id <document_id> --table-id <table_block_id> --row 0 --col 0

# 表格属性
feishu-docs-cli doc update table props --doc-id <document_id> --table-id <table_block_id> --header-row=true
feishu-docs-cli doc update table props --doc-id <document_id> --table-id <table_block_id> --column-index 0 --column-width 240
```

### 知识库(Wiki)操作

```bash
# 列出知识空间
feishu-docs-cli wiki spaces

# 获取知识空间信息
feishu-docs-cli wiki space --id <space_id>

# 获取节点信息
feishu-docs-cli wiki node --token <node_token>

# 解析 wiki 链接获取实际文档
feishu-docs-cli wiki resolve --url "https://xxx.feishu.cn/wiki/ABC123"
feishu-docs-cli wiki resolve --url "wiki/ABC123"

# 列出子节点
feishu-docs-cli wiki nodes --space-id <space_id> [--parent <parent_node_token>]

# 创建节点
feishu-docs-cli wiki create --space-id <space_id> --title "新页面" [--obj-type docx] [--parent <parent_token>]

# 移动节点
feishu-docs-cli wiki move --space-id <space_id> --node-token <token> --target-parent <target_token>

# 复制节点
feishu-docs-cli wiki copy --space-id <space_id> --node-token <token> --target-parent <target_token> --target-space <target_space_id>
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
