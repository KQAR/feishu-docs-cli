# feishu-docs-cli CLI 技能

基于 [KQAR/feishu-docs-cli](https://github.com/KQAR/feishu-docs-cli) CLI 工具的飞书文档操作技能。

## 功能特性

- ✅ **文档 CRUD**：创建、读取、更新、删除文档
- ✅ **统一写操作**：append（追加 Markdown）、insert（插入块）、set-text（更新块）、delete（删除块）
- ✅ **表格管理**：create/show/write/set-cell/insert-row/insert-column/delete-rows/delete-columns/merge/unmerge/props
- ✅ **Wiki 解析**：自动解析 wiki 链接获取实际文档类型和 token
- ✅ **Markdown 支持**：标题、列表、代码块、分割线、引用、待办（仅块级语法）
- ✅ **知识库管理**：空间、节点的完整操作

## ⚠️ Markdown append 的重要限制

`doc update append --markdown` 仅支持块级 Markdown 语法。以下语法不会被解析：

| 语法 | 结果 | 替代方案 |
|---|---|---|
| `\| col \| col \|` 表格 | 作为带 `\|` 的纯文本显示 | 用 `doc update table create --data` 创建真正的表格 |
| `**bold**` 加粗 | 星号作为字面字符显示 | 用 `####` 四级标题代替，或直接不用加粗 |
| `[text](url)` 链接 | 作为字面字符显示 | 目前无 CLI 替代方案 |
| `` `code` `` 行内代码 | 反引号作为字面字符显示 | 目前无 CLI 替代方案 |

### 正确的写文档模式

文本用 `append`，表格用 `table create`，交替调用：

```bash
# 1. 写文本
cat << 'MD' | feishu-docs-cli doc update append -d $DOC --markdown -
## 章节标题
这里是正文。
MD

# 2. 创建表格（TSV 格式）
echo '姓名\t年龄\n张三\t25' | feishu-docs-cli doc update table create -d $DOC --header-row --data -

# 3. 继续写文本
cat << 'MD' | feishu-docs-cli doc update append -d $DOC --markdown -
## 下一章节
更多内容...
MD
```

### 单次 append 大小限制

飞书 API 限制单次约 50 个块。内容过多会报 `99992402: field validation failed`。解决：拆分为多次 `append` 调用。

## 安装

### 下载预编译二进制

```bash
# Linux ARM64 (iSH 环境)
curl -sL -o /usr/local/bin/feishu-docs-cli \
  "https://github.com/KQAR/feishu-docs-cli/releases/download/v0.2.0/feishu-docs-cli_0.2.0_linux_arm64.tar.gz"
tar -xzf /usr/local/bin/feishu-docs-cli -C /usr/local/bin/
chmod +x /usr/local/bin/feishu-docs-cli
```

### 配置

```bash
# 初始化配置
feishu-docs-cli init

# 编辑配置文件
# 路径: ~/.config/feishu-docs/config.json
cat > ~/.config/feishu-docs/config.json << EOF
{
  "app_id": "YOUR_APP_ID",
  "app_secret": "YOUR_APP_SECRET"
}
EOF
```

## 使用示例

### 文档操作

```bash
# 创建文档
feishu-docs-cli doc create --title "AI周报" --folder-token Fldxxx

# 获取文档内容
feishu-docs-cli doc content --id <doc_id>

# 列出文档块
feishu-docs-cli doc blocks --id <doc_id>

# 追加 Markdown（从 stdin 管道读取）
cat content.md | feishu-docs-cli doc update append -d <doc_id> --markdown -

# 插入单个块
feishu-docs-cli doc update insert -d <doc_id> --text "新段落" --type text
feishu-docs-cli doc update insert -d <doc_id> --text "标题" --type heading2

# 更新块文本
feishu-docs-cli doc update set-text -d <doc_id> -b <block_id> --text "新内容"

# 删除块范围
feishu-docs-cli doc update delete -d <doc_id> -b <parent_block_id> --start 0 --end 5
```

### 表格操作

```bash
# 创建表格（TSV 格式，Tab 分隔）
echo '姓名\t角色\n张三\tOwner' | feishu-docs-cli doc update table create -d <doc_id> --header-row --data -

# 创建表格（JSON 矩阵格式）
feishu-docs-cli doc update table create -d <doc_id> --data '[["姓名","角色"],["张三","Owner"]]' --header-row

# 查看表格
feishu-docs-cli doc update table show -d <doc_id> -t <table_block_id> --format table

# 重写整张表格
feishu-docs-cli doc update table write -d <doc_id> -t <table_block_id> --data $'A\tB\n1\t2'

# 更新单元格
feishu-docs-cli doc update table set-cell -d <doc_id> -t <table_block_id> --row 1 --col 0 --text "李四"

# 合并单元格
feishu-docs-cli doc update table merge -d <doc_id> -t <table_block_id> --row-start 0 --row-end 1 --column-start 0 --column-end 2
```

### Wiki 操作

```bash
# 解析 wiki 链接
feishu-docs-cli wiki resolve -u "https://xxx.feishu.cn/wiki/ABC123"

# 创建 wiki 节点
feishu-docs-cli wiki create -s SPACE_ID -t "新页面" -p PARENT_NODE_TOKEN

# 列出子节点
feishu-docs-cli wiki nodes -s SPACE_ID [-p PARENT_NODE_TOKEN]
```

## 权限要求

请确保飞书应用已开通以下权限：

- `docx:document` - 文档读写权限
- `wiki:wiki` - 知识库读写权限

## 参考

- 飞书官方 OpenClaw 插件: https://github.com/larksuite/openclaw-lark
- 飞书开放平台文档: https://open.feishu.cn/
