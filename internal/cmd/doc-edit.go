package cmd

import (
	"context"

	"github.com/spf13/cobra"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"

	"github.com/KQAR/feishu-docs/internal/output"
)

func newDocInsertBlockCmd() *cobra.Command {
	var documentID, blockID, text string
	var index int
	var blockType string

	cmd := &cobra.Command{
		Use:   "insert",
		Short: "向文档中插入内容块",
		Long:  "在指定块下插入新的子块。block-id 默认使用文档 ID（即插入到文档根节点下）。",
		Run: func(cmd *cobra.Command, args []string) {
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

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "父块 ID (默认为文档根节点)")
	cmd.Flags().StringVarP(&text, "text", "t", "", "文本内容 (必填)")
	cmd.Flags().IntVar(&index, "index", -1, "插入位置索引 (-1 表示末尾)")
	cmd.Flags().StringVar(&blockType, "type", "text", "块类型: text/heading1/heading2/heading3/heading4/heading5/heading6/heading7/heading8/heading9/bullet/ordered/code/todo")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("text")

	return cmd
}

func newDocUpdateBlockCmd() *cobra.Command {
	var documentID, blockID, text string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "更新文档块内容",
		Long:  "替换指定文本块的内容。仅支持文本类块（text、heading 等）。",
		Run: func(cmd *cobra.Command, args []string) {
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

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "块 ID (必填)")
	cmd.Flags().StringVarP(&text, "text", "t", "", "新的文本内容 (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("block-id")
	cmd.MarkFlagRequired("text")

	return cmd
}

func newDocGetBlockCmd() *cobra.Command {
	var documentID, blockID string

	cmd := &cobra.Command{
		Use:   "block",
		Short: "获取单个块的详细信息",
		Run: func(cmd *cobra.Command, args []string) {
			req := larkdocx.NewGetDocumentBlockReqBuilder().
				DocumentId(documentID).
				BlockId(blockID).
				Build()

			resp, err := larkClient.Docx.DocumentBlock.Get(context.Background(), req)
			if err != nil {
				output.Errorf("获取块信息失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取块信息失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "块 ID (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("block-id")

	return cmd
}

// buildTextBlock 根据类型和内容构建文本 Block
func buildTextBlock(content string, blockType string) *larkdocx.Block {
	textBody := larkdocx.NewTextBuilder().
		Elements([]*larkdocx.TextElement{
			larkdocx.NewTextElementBuilder().
				TextRun(larkdocx.NewTextRunBuilder().Content(content).Build()).
				Build(),
		}).
		Build()

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
