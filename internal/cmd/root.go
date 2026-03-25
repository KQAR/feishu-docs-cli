package cmd

import (
	"fmt"
	"os"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/cobra"

	"github.com/KQAR/feishu-docs-cli/internal/client"
	"github.com/KQAR/feishu-docs-cli/internal/config"
	"github.com/KQAR/feishu-docs-cli/internal/output"
	"github.com/KQAR/feishu-docs-cli/internal/version"
)

var larkClient *lark.Client

var rootCmd = &cobra.Command{
	Use:   "feishu-docs-cli",
	Short: "飞书文档与知识库 CLI 管理工具",
	Long: `飞书文档与知识库 CLI 管理工具 (AI-Friendly)

基于飞书开放平台 Golang SDK 的命令行工具，支持文档和知识库(Wiki)的增删改查操作。
所有 --doc-id / -d / -i 参数均可直接传入 wiki 链接，自动解析为文档 ID。

## 快速开始

  feishu-docs-cli init                     # 初始化配置（填入 app_id / app_secret）
  feishu-docs-cli doc content -i <DOC>     # 读取文档纯文本
  feishu-docs-cli doc blocks -i <DOC>      # 列出所有块（获取 block_id 和结构）

## 写文档的核心模式

  文本用 append，表格用 table create，交替调用：

  # 写标题和段落
  cat <<'MD' | feishu-docs-cli doc update append -d <DOC> --markdown -
  # 标题
  正文段落
  - 列表项
  MD

  # 创建表格（TSV 格式，Tab 分隔列）
  printf '列A\t列B\n值1\t值2' | feishu-docs-cli doc update table create -d <DOC> --header-row --data -

## 关键限制（踩坑必读）

  1. append 只支持块级 Markdown，不支持行内样式：
     **加粗** → 字面显示星号    *斜体* → 字面显示星号
     [链接](url) → 字面显示     | 表格 | → 字面显示
     替代方案：用 #### 四级标题代替加粗；用 table create 代替 Markdown 表格

  2. 单次 append 上限约 50 个块，超出报 99992402 错误
     解决：拆分为多次 append 调用，每次 ≤ 45 块

  3. 内置 API 限流保护（300ms 间隔 + 429 自动重试），无需手动 sleep

## 完整重写文档

  # 1. 获取子块总数
  feishu-docs-cli doc blocks -i <DOC> | python3 -c "
    import json,sys; d=json.load(sys.stdin)
    print(len(d['items'][0].get('children',[])))
  "
  # 2. 删除所有子块（block-id 用 items[0].block_id，不能用 wiki 链接）
  feishu-docs-cli doc update delete -d <DOC> -b <ROOT_BLOCK_ID> --start 0 --end <总数>
  # 3. 按顺序重建（append + table create 交替）`,
	Version: version.Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "init" || cmd.Name() == "version" {
			return
		}

		cfg, err := config.Load()
		if err != nil {
			output.Errorf("加载配置失败: %v\n请先运行 `feishu-docs-cli init` 初始化配置", err)
		}
		larkClient = client.New(cfg)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Long:  "在 ~/.config/feishu-docs-cli/config.json 创建配置模板，需手动填入 app_id 和 app_secret。",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.EnsureConfigFile()
		if err != nil {
			output.Errorf("初始化配置失败: %v", err)
		}
		fmt.Printf("配置文件路径: %s\n请编辑该文件，填入飞书应用的 app_id 和 app_secret。\n", path)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示详细版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Full())
	},
}

// Execute 执行 CLI 入口
func Execute() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(newDocCmd())
	rootCmd.AddCommand(newWikiCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
