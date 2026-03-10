package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"

	"github.com/KQAR/feishu-docs/internal/output"
)

func newDocCmd() *cobra.Command {
	docCmd := &cobra.Command{
		Use:   "doc",
		Short: "文档管理",
		Long:  "管理飞书新版文档(Docx)，支持创建、查看、插入、编辑、删除等操作。",
	}

	docCmd.AddCommand(
		newDocCreateCmd(),
		newDocGetCmd(),
		newDocContentCmd(),
		newDocListBlocksCmd(),
		newDocGetBlockCmd(),
		newDocInsertBlockCmd(),
		newDocUpdateBlockCmd(),
		newDocDeleteBlocksCmd(),
	)

	return docCmd
}

func newDocCreateCmd() *cobra.Command {
	var title, folderToken string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建文档",
		Run: func(cmd *cobra.Command, args []string) {
			bodyBuilder := larkdocx.NewCreateDocumentReqBodyBuilder()
			if title != "" {
				bodyBuilder.Title(title)
			}
			if folderToken != "" {
				bodyBuilder.FolderToken(folderToken)
			}

			req := larkdocx.NewCreateDocumentReqBuilder().
				Body(bodyBuilder.Build()).
				Build()

			resp, err := larkClient.Docx.Document.Create(context.Background(), req)
			if err != nil {
				output.Errorf("创建文档失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("创建文档失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("文档创建成功")
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "文档标题")
	cmd.Flags().StringVarP(&folderToken, "folder-token", "f", "", "目标文件夹 token")

	return cmd
}

func newDocGetCmd() *cobra.Command {
	var documentID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "获取文档信息",
		Run: func(cmd *cobra.Command, args []string) {
			req := larkdocx.NewGetDocumentReqBuilder().
				DocumentId(documentID).
				Build()

			resp, err := larkClient.Docx.Document.Get(context.Background(), req)
			if err != nil {
				output.Errorf("获取文档失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取文档失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "id", "i", "", "文档 ID (必填)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newDocContentCmd() *cobra.Command {
	var documentID string
	var lang int

	cmd := &cobra.Command{
		Use:   "content",
		Short: "获取文档纯文本内容",
		Run: func(cmd *cobra.Command, args []string) {
			req := larkdocx.NewRawContentDocumentReqBuilder().
				DocumentId(documentID).
				Lang(lang).
				Build()

			resp, err := larkClient.Docx.Document.RawContent(context.Background(), req)
			if err != nil {
				output.Errorf("获取文档内容失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取文档内容失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			if resp.Data != nil && resp.Data.Content != nil {
				fmt.Println(*resp.Data.Content)
			}
		},
	}

	cmd.Flags().StringVarP(&documentID, "id", "i", "", "文档 ID (必填)")
	cmd.Flags().IntVar(&lang, "lang", 0, "语言 (0: 中文, 1: 英文, 2: 日文)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newDocListBlocksCmd() *cobra.Command {
	var documentID, pageToken string
	var pageSize int

	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "列出文档所有块",
		Run: func(cmd *cobra.Command, args []string) {
			builder := larkdocx.NewListDocumentBlockReqBuilder().
				DocumentId(documentID).
				PageSize(pageSize)

			if pageToken != "" {
				builder.PageToken(pageToken)
			}

			req := builder.Build()

			resp, err := larkClient.Docx.DocumentBlock.List(context.Background(), req)
			if err != nil {
				output.Errorf("获取文档块列表失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取文档块列表失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "id", "i", "", "文档 ID (必填)")
	cmd.Flags().IntVarP(&pageSize, "page-size", "n", 50, "每页数量")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "分页标记")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newDocDeleteBlocksCmd() *cobra.Command {
	var documentID, blockID string
	var startIndex, endIndex int

	cmd := &cobra.Command{
		Use:   "delete-blocks",
		Short: "删除文档子块",
		Long:  "删除指定块下指定范围的子块。",
		Run: func(cmd *cobra.Command, args []string) {
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

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "父块 ID (必填)")
	cmd.Flags().IntVar(&startIndex, "start", 0, "起始索引 (必填)")
	cmd.Flags().IntVar(&endIndex, "end", 0, "结束索引 (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("block-id")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}
