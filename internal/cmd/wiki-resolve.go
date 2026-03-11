package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/KQAR/feishu-docs/internal/output"
)

// newWikiResolveCmd 解析 wiki URL/token 获取实际文档信息
func newWikiResolveCmd() *cobra.Command {
	var wikiURL string

	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "解析 wiki 链接获取实际文档",
		Long: `解析知识库链接（/wiki/TOKEN）获取实际文档类型和 token。

知识库链接背后可能是云文档、电子表格、多维表格等不同类型文档。
此命令会先查询实际类型，返回 obj_type 和 obj_token。

支持的 URL 格式：
  - https://xxx.feishu.cn/wiki/TOKEN
  - wiki/TOKEN
  - 纯 TOKEN

返回信息：
  - obj_type: 文档类型 (docx/sheet/bitable/etc)
  - obj_token: 实际文档 token
  - title: 文档标题`,
		Run: func(cmd *cobra.Command, args []string) {
			token := extractWikiToken(wikiURL)
			if token == "" {
				output.Errorf("无法从 URL 提取 wiki token: %s", wikiURL)
			}

			req := larkwiki.NewGetNodeSpaceReqBuilder().
				Token(token).
				Build()

			resp, err := larkClient.Wiki.Space.GetNode(context.Background(), req)
			if err != nil {
				output.Errorf("获取 wiki 节点信息失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取 wiki 节点信息失败 [%d]: %s (request_id: %s)",
					resp.Code, resp.Msg, resp.RequestId())
			}

			if resp.Data == nil || resp.Data.Node == nil {
				output.Errorf("wiki 节点不存在或无权访问")
			}

			node := resp.Data.Node

			// 输出结果
			fmt.Println("=== Wiki 节点解析结果 ===")
			fmt.Printf("Wiki Token:    %s\n", token)
			fmt.Printf("节点标题:      %s\n", deref(node.Title))
			fmt.Printf("Obj Type:      %s\n", deref(node.ObjType))
			fmt.Printf("Obj Token:     %s\n", deref(node.ObjToken))
			fmt.Printf("节点类型:      %s\n", deref(node.NodeType))
			fmt.Printf("父节点 Token:  %s\n", deref(node.ParentNodeToken))
			fmt.Printf("空间 ID:       %s\n", deref(node.SpaceId))

			// JSON 输出
			output.JSON(map[string]interface{}{
				"wiki_token":   token,
				"title":        deref(node.Title),
				"obj_type":     deref(node.ObjType),
				"obj_token":    deref(node.ObjToken),
				"node_type":    deref(node.NodeType),
				"parent_node_token": deref(node.ParentNodeToken),
				"space_id":     deref(node.SpaceId),
			})

			// 根据类型给出建议
			fmt.Println("\n=== 后续操作建议 ===")
			switch deref(node.ObjType) {
			case "docx":
				fmt.Printf("文档类型: 云文档 (docx)\n")
				fmt.Printf("使用命令: feishu-docs doc get --id %s\n", deref(node.ObjToken))
			case "sheet":
				fmt.Printf("文档类型: 电子表格 (sheet)\n")
				fmt.Printf("使用命令: feishu-docs sheet get --id %s (暂未实现)\n", deref(node.ObjToken))
			case "bitable":
				fmt.Printf("文档类型: 多维表格 (bitable)\n")
				fmt.Printf("使用命令: feishu-docs bitable get --id %s (暂未实现)\n", deref(node.ObjToken))
			case "mindnote":
				fmt.Printf("文档类型: 思维笔记 (mindnote)\n")
				fmt.Printf("暂不支持此类型文档\n")
			default:
				fmt.Printf("文档类型: %s (未知类型)\n", deref(node.ObjToken))
				fmt.Printf("暂不支持此类型文档\n")
			}
		},
	}

	cmd.Flags().StringVarP(&wikiURL, "url", "u", "", "wiki URL 或 token (必填)")
	cmd.MarkFlagRequired("url")

	return cmd
}

// extractWikiToken 从各种格式的输入中提取 wiki token
func extractWikiToken(input string) string {
	input = strings.TrimSpace(input)

	// 匹配 https://xxx.feishu.cn/wiki/TOKEN
	re1 := regexp.MustCompile(`/wiki/([a-zA-Z0-9]+)`)
	if matches := re1.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	// 匹配 wiki/TOKEN
	re2 := regexp.MustCompile(`^wiki/([a-zA-Z0-9]+)$`)
	if matches := re2.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	// 纯 TOKEN (字母数字组合)
	re3 := regexp.MustCompile(`^([a-zA-Z0-9]+)$`)
	if matches := re3.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	return ""
}


