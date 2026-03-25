package cmd

import (
	"context"

	"github.com/spf13/cobra"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"

	"github.com/KQAR/feishu-docs-cli/internal/output"
)

func newDocGetBlockCmd() *cobra.Command {
	var documentID, blockID string

	cmd := &cobra.Command{
		Use:   "block",
		Short: "获取单个块的详细信息",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
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

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "块 ID (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("block-id")

	return cmd
}
