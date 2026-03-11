package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/KQAR/feishu-docs/internal/output"
)

func newDocUpdateV2Cmd() *cobra.Command {
	var documentID, markdown string

	cmd := &cobra.Command{
		Use:   "update-v2",
		Short: "批量追加 Markdown 内容到文档",
		Long: `将 Markdown 内容解析为飞书文档块并追加到文档末尾。

支持的 Markdown 语法:
  # 标题1
  ## 标题2
  ### 标题3
  - 无序列表
  * 无序列表
  1. 有序列表
  [ ] 待办事项
  [x] 已完成待办
  普通文本

示例:
  feishu-docs doc update-v2 --doc-id DOC_ID --markdown "# 新标题\n\n正文内容"`,
		Run: func(cmd *cobra.Command, args []string) {
			blocks := parseMarkdownToBlocks(markdown)

			req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
				DocumentId(documentID).
				BlockId(documentID).
				DocumentRevisionId(-1).
				Body(larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
					Children(blocks).
					Index(-1).
					Build()).
				Build()

			resp, err := larkClient.Docx.DocumentBlockChildren.Create(context.Background(), req)
			if err != nil {
				output.Errorf("追加内容失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("追加内容失败 [%d]: %s", resp.Code, resp.Msg)
			}

			output.Success(fmt.Sprintf("成功追加 %d 个内容块", len(blocks)))
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID (必填)")
	cmd.Flags().StringVar(&markdown, "markdown", "", "Markdown 内容 (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("markdown")

	return cmd
}

func parseMarkdownToBlocks(markdown string) []*larkdocx.Block {
	var blocks []*larkdocx.Block
	lines := strings.Split(markdown, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		block := parseLineToBlock(line)
		if block != nil {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func parseLineToBlock(line string) *larkdocx.Block {
	textBody := larkdocx.NewTextBuilder().
		Elements([]*larkdocx.TextElement{
			larkdocx.NewTextElementBuilder().
				TextRun(larkdocx.NewTextRunBuilder().Content(line).Build()).
				Build(),
		}).
		Build()

	builder := larkdocx.NewBlockBuilder()

	if strings.HasPrefix(line, "# ") {
		builder.BlockType(3).Heading1(textBody)
	} else if strings.HasPrefix(line, "## ") {
		builder.BlockType(4).Heading2(textBody)
	} else if strings.HasPrefix(line, "### ") {
		builder.BlockType(5).Heading3(textBody)
	} else if strings.HasPrefix(line, "#### ") {
		builder.BlockType(6).Heading4(textBody)
	} else if strings.HasPrefix(line, "##### ") {
		builder.BlockType(7).Heading5(textBody)
	} else if strings.HasPrefix(line, "###### ") {
		builder.BlockType(8).Heading6(textBody)
	} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		builder.BlockType(12).Bullet(textBody)
	} else if regexp.MustCompile(`^\d+\.\s`).MatchString(line) {
		builder.BlockType(13).Ordered(textBody)
	} else if strings.HasPrefix(line, "[ ] ") || strings.HasPrefix(line, "[x] ") {
		builder.BlockType(17).Todo(textBody)
	} else {
		builder.BlockType(2).Text(textBody)
	}

	return builder.Build()
}
