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
	Use:     "feishu-docs-cli",
	Short:   "飞书文档与知识库 CLI 管理工具",
	Long:    "基于飞书开放平台 Golang SDK 的命令行工具，支持文档和知识库(Wiki)的增删改查操作。",
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
