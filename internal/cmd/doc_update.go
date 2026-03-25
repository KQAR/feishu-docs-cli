package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"

	"github.com/KQAR/feishu-docs-cli/internal/output"
)

var (
	reOrderedList = regexp.MustCompile(`^\d+\.\s`)
	reOrderedTrim = regexp.MustCompile(`^\d+\.\s+`)
	reCodeFence   = regexp.MustCompile("^`{3,}")
)

func newDocUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "统一文档写操作",
		Long:  "统一管理文档写操作，包括 Markdown 追加、文本块插入与更新、块删除以及表格编辑。",
	}

	updateCmd.AddCommand(
		newDocUpdateAppendCmd(),
		newDocUpdateInsertCmd(),
		newDocUpdateSetTextCmd(),
		newDocUpdateDeleteCmd(),
		newDocUpdateTableCmd(),
	)

	return updateCmd
}

func newDocUpdateAppendCmd() *cobra.Command {
	var documentID, blockID, markdown string
	var index int

	cmd := &cobra.Command{
		Use:   "append",
		Short: "追加 Markdown 块",
		Long: `将 Markdown 内容解析为飞书文档块并写入指定父块。

支持的 Markdown 语法:
  # 标题1
  ## 标题2
  ### 标题3
  - 无序列表
  * 无序列表
  1. 有序列表
  [ ] 待办事项
  [x] 已完成待办
  普通文本`,
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if blockID == "" {
				blockID = documentID
			}

			var err error
			if markdown, err = readStdinOrValue(markdown); err != nil {
				output.Errorf("%v", err)
			}

			blocks := parseMarkdownToBlocks(markdown)
			if len(blocks) == 0 {
				output.Errorf("没有可写入的 Markdown 内容")
			}

			req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
				DocumentId(documentID).
				BlockId(blockID).
				DocumentRevisionId(-1).
				Body(larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
					Children(blocks).
					Index(index).
					Build()).
				Build()

			resp, err := larkClient.Docx.DocumentBlockChildren.Create(context.Background(), req)
			if err != nil {
				output.Errorf("追加 Markdown 失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("追加 Markdown 失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success(fmt.Sprintf("成功写入 %d 个 Markdown 块", len(blocks)))
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "父块 ID (默认文档根节点)")
	cmd.Flags().StringVar(&markdown, "markdown", "", "Markdown 内容，- 表示从 stdin 读取 (必填)")
	cmd.Flags().IntVar(&index, "index", -1, "插入位置索引 (-1 表示末尾)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("markdown")

	return cmd
}

func newDocUpdateInsertCmd() *cobra.Command {
	var documentID, blockID, text string
	var index int
	var blockType string

	cmd := &cobra.Command{
		Use:   "insert",
		Short: "插入单个文本块",
		Long:  "在指定父块下插入单个文本类块。block-id 默认使用文档 ID（即文档根节点）。",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if blockID == "" {
				blockID = documentID
			}

			block := buildTextBlock(text, blockType)

			req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
				DocumentId(documentID).
				BlockId(blockID).
				DocumentRevisionId(-1).
				Body(larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
					Children([]*larkdocx.Block{block}).
					Index(index).
					Build()).
				Build()

			resp, err := larkClient.Docx.DocumentBlockChildren.Create(context.Background(), req)
			if err != nil {
				output.Errorf("插入内容失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("插入内容失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("内容插入成功")
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "父块 ID (默认为文档根节点)")
	cmd.Flags().StringVarP(&text, "text", "t", "", "文本内容 (必填)")
	cmd.Flags().IntVar(&index, "index", -1, "插入位置索引 (-1 表示末尾)")
	cmd.Flags().StringVar(&blockType, "type", "text", "块类型: text/heading1/heading2/heading3/heading4/heading5/heading6/heading7/heading8/heading9/bullet/ordered/code/todo")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("text")

	return cmd
}

func newDocUpdateSetTextCmd() *cobra.Command {
	var documentID, blockID, text string

	cmd := &cobra.Command{
		Use:   "set-text",
		Short: "更新文本块内容",
		Long:  "替换指定文本块的内容。仅支持文本类块（text、heading 等）。",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			elements := []*larkdocx.TextElement{
				larkdocx.NewTextElementBuilder().
					TextRun(larkdocx.NewTextRunBuilder().
						Content(text).
						Build()).
					Build(),
			}

			req := larkdocx.NewPatchDocumentBlockReqBuilder().
				DocumentId(documentID).
				BlockId(blockID).
				DocumentRevisionId(-1).
				UpdateBlockRequest(larkdocx.NewUpdateBlockRequestBuilder().
					UpdateTextElements(larkdocx.NewUpdateTextElementsRequestBuilder().
						Elements(elements).
						Build()).
					Build()).
				Build()

			resp, err := larkClient.Docx.DocumentBlock.Patch(context.Background(), req)
			if err != nil {
				output.Errorf("更新块内容失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("更新块内容失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("块内容更新成功")
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "块 ID (必填)")
	cmd.Flags().StringVarP(&text, "text", "t", "", "新的文本内容 (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("block-id")
	cmd.MarkFlagRequired("text")

	return cmd
}

func newDocUpdateDeleteCmd() *cobra.Command {
	var documentID, blockID string
	var startIndex, endIndex int

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除指定范围的子块",
		Long:  "删除指定父块下指定范围的子块。",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			req := larkdocx.NewBatchDeleteDocumentBlockChildrenReqBuilder().
				DocumentId(documentID).
				BlockId(blockID).
				Body(larkdocx.NewBatchDeleteDocumentBlockChildrenReqBodyBuilder().
					StartIndex(startIndex).
					EndIndex(endIndex).
					Build()).
				Build()

			resp, err := larkClient.Docx.DocumentBlockChildren.BatchDelete(context.Background(), req)
			if err != nil {
				output.Errorf("删除块失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("删除块失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("块删除成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "父块 ID (必填)")
	cmd.Flags().IntVar(&startIndex, "start", 0, "起始索引 (左闭右开)")
	cmd.Flags().IntVar(&endIndex, "end", 0, "结束索引 (左闭右开)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("block-id")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func buildTextBlock(content string, blockType string) *larkdocx.Block {
	textBody := newTextBody(content)
	builder := larkdocx.NewBlockBuilder()

	switch blockType {
	case "heading1":
		builder.BlockType(3).Heading1(textBody)
	case "heading2":
		builder.BlockType(4).Heading2(textBody)
	case "heading3":
		builder.BlockType(5).Heading3(textBody)
	case "heading4":
		builder.BlockType(6).Heading4(textBody)
	case "heading5":
		builder.BlockType(7).Heading5(textBody)
	case "heading6":
		builder.BlockType(8).Heading6(textBody)
	case "heading7":
		builder.BlockType(9).Heading7(textBody)
	case "heading8":
		builder.BlockType(10).Heading8(textBody)
	case "heading9":
		builder.BlockType(11).Heading9(textBody)
	case "bullet":
		builder.BlockType(12).Bullet(textBody)
	case "ordered":
		builder.BlockType(13).Ordered(textBody)
	case "code":
		builder.BlockType(14).Code(textBody)
	case "todo":
		builder.BlockType(17).Todo(textBody)
	default:
		builder.BlockType(2).Text(textBody)
	}

	return builder.Build()
}

func parseMarkdownToBlocks(markdown string) []*larkdocx.Block {
	var blocks []*larkdocx.Block
	lines := strings.Split(markdown, "\n")

	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])

		// Fenced code block (``` ... ```)
		if reCodeFence.MatchString(trimmed) {
			fence := reCodeFence.FindString(trimmed)
			var codeLines []string
			i++
			for i < len(lines) {
				if strings.TrimSpace(lines[i]) == fence {
					i++
					break
				}
				codeLines = append(codeLines, lines[i])
				i++
			}
			blocks = append(blocks, larkdocx.NewBlockBuilder().
				BlockType(14).
				Code(newTextBody(strings.Join(codeLines, "\n"))).
				Build())
			continue
		}

		// Skip empty lines
		if trimmed == "" {
			i++
			continue
		}

		// Horizontal rule (---, ***, ___)
		if isHorizontalRule(trimmed) {
			blocks = append(blocks, larkdocx.NewBlockBuilder().
				BlockType(22).
				Divider(larkdocx.NewDividerBuilder().Build()).
				Build())
			i++
			continue
		}

		// Blockquote (> ...) — parsed as regular text blocks
		if strings.HasPrefix(trimmed, "> ") || trimmed == ">" {
			var quoteLines []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if strings.HasPrefix(t, "> ") {
					quoteLines = append(quoteLines, strings.TrimPrefix(t, "> "))
					i++
				} else if t == ">" {
					quoteLines = append(quoteLines, "")
					i++
				} else {
					break
				}
			}
			for _, ql := range quoteLines {
				if ql != "" {
					blocks = append(blocks, parseLineToBlock(ql))
				}
			}
			continue
		}

		// Standard single-line block
		block := parseLineToBlock(trimmed)
		if block != nil {
			blocks = append(blocks, block)
		}
		i++
	}

	return blocks
}

func parseLineToBlock(line string) *larkdocx.Block {
	builder := larkdocx.NewBlockBuilder()
	content := line

	if strings.HasPrefix(line, "# ") {
		content = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		builder.BlockType(3).Heading1(newTextBody(content))
	} else if strings.HasPrefix(line, "## ") {
		content = strings.TrimSpace(strings.TrimPrefix(line, "## "))
		builder.BlockType(4).Heading2(newTextBody(content))
	} else if strings.HasPrefix(line, "### ") {
		content = strings.TrimSpace(strings.TrimPrefix(line, "### "))
		builder.BlockType(5).Heading3(newTextBody(content))
	} else if strings.HasPrefix(line, "#### ") {
		content = strings.TrimSpace(strings.TrimPrefix(line, "#### "))
		builder.BlockType(6).Heading4(newTextBody(content))
	} else if strings.HasPrefix(line, "##### ") {
		content = strings.TrimSpace(strings.TrimPrefix(line, "##### "))
		builder.BlockType(7).Heading5(newTextBody(content))
	} else if strings.HasPrefix(line, "###### ") {
		content = strings.TrimSpace(strings.TrimPrefix(line, "###### "))
		builder.BlockType(8).Heading6(newTextBody(content))
	} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		content = strings.TrimSpace(line[2:])
		builder.BlockType(12).Bullet(newTextBody(content))
	} else if reOrderedList.MatchString(line) {
		content = reOrderedTrim.ReplaceAllString(line, "")
		builder.BlockType(13).Ordered(newTextBody(content))
	} else if strings.HasPrefix(line, "[ ] ") || strings.HasPrefix(line, "[x] ") {
		content = strings.TrimSpace(line[4:])
		builder.BlockType(17).Todo(newTextBody(content))
	} else {
		builder.BlockType(2).Text(newTextBody(content))
	}

	return builder.Build()
}

func isHorizontalRule(line string) bool {
	stripped := strings.ReplaceAll(line, " ", "")
	if len(stripped) < 3 {
		return false
	}
	for _, ch := range []byte{'-', '*', '_'} {
		if strings.Count(stripped, string(ch)) == len(stripped) {
			return true
		}
	}
	return false
}

func readStdinOrValue(value string) (string, error) {
	if value != "-" {
		return value, nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("读取 stdin 失败: %w", err)
	}
	return string(data), nil
}
