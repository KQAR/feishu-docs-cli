package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"

	"github.com/KQAR/feishu-docs-cli/internal/output"
)

const (
	docxBlockTypeText      = 2
	docxBlockTypeTable     = 31
	docxBlockTypeTableCell = 32
)

type tableSnapshot struct {
	Table        *larkdocx.Block
	Blocks       map[string]*larkdocx.Block
	Rows         int
	Cols         int
	CellIDs      []string
	HeaderRow    bool
	HeaderColumn bool
	ColumnWidths []int
	MergeInfo    []*larkdocx.TableMergeInfo
}

func newDocUpdateTableCmd() *cobra.Command {
	tableCmd := &cobra.Command{
		Use:   "table",
		Short: "表格构建与编辑",
		Long:  "管理飞书文档中的表格，支持建表、查看、写入单元格、插入/删除行列和合并操作。",
	}

	tableCmd.AddCommand(
		newDocTableCreateCmd(),
		newDocTableShowCmd(),
		newDocTableWriteCmd(),
		newDocTableSetCellCmd(),
		newDocTableInsertRowCmd(),
		newDocTableInsertColumnCmd(),
		newDocTableDeleteRowsCmd(),
		newDocTableDeleteColumnsCmd(),
		newDocTableMergeCmd(),
		newDocTableUnmergeCmd(),
		newDocTablePropsCmd(),
	)

	return tableCmd
}

func newDocTableCreateCmd() *cobra.Command {
	var documentID, blockID, data, dataFile, columnWidths string
	var rows, cols, index int
	var headerRow, headerColumn bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建表格",
		Long: `创建文档表格，并可选地一次性填充单元格内容。

数据格式默认使用 TSV：
  A1\tB1
  A2\tB2

也支持 JSON 矩阵：
  [["A1","B1"],["A2","B2"]]`,
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if blockID == "" {
				blockID = documentID
			}

			matrix, err := parseTableMatrixInput(data, dataFile, rows, cols)
			if err != nil {
				output.Errorf("解析表格数据失败: %v", err)
			}

			widths, err := parseColumnWidths(columnWidths, len(matrix[0]))
			if err != nil {
				output.Errorf("解析列宽失败: %v", err)
			}

			tempTableID, descendants := buildTableDescendants(blockID, matrix, widths, headerRow, headerColumn)

			bodyBuilder := larkdocx.NewCreateDocumentBlockDescendantReqBodyBuilder().
				ChildrenId([]string{tempTableID}).
				Descendants(descendants)
			if index >= 0 {
				bodyBuilder.Index(index)
			}

			req := larkdocx.NewCreateDocumentBlockDescendantReqBuilder().
				DocumentId(documentID).
				BlockId(blockID).
				DocumentRevisionId(-1).
				Body(bodyBuilder.Build()).
				Build()

			resp, err := larkClient.Docx.DocumentBlockDescendant.Create(context.Background(), req)
			if err != nil {
				output.Errorf("创建表格失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("创建表格失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success(fmt.Sprintf("表格创建成功 (%d 行 x %d 列)", len(matrix), len(matrix[0])))
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&blockID, "block-id", "b", "", "父块 ID (默认文档根节点)")
	cmd.Flags().IntVar(&rows, "rows", 0, "表格行数；若提供 data 且未指定则自动推断")
	cmd.Flags().IntVar(&cols, "cols", 0, "表格列数；若提供 data 且未指定则自动推断")
	cmd.Flags().IntVar(&index, "index", -1, "插入位置索引；默认追加到末尾")
	cmd.Flags().StringVar(&data, "data", "", "表格数据，支持 TSV/JSON 矩阵，- 表示从 stdin 读取")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "表格数据文件路径，支持 TSV 或 JSON 矩阵")
	cmd.Flags().StringVar(&columnWidths, "column-widths", "", "列宽列表，逗号分隔，例如 240,180,180")
	cmd.Flags().BoolVar(&headerRow, "header-row", false, "设置首行为标题行")
	cmd.Flags().BoolVar(&headerColumn, "header-column", false, "设置首列为标题列")
	cmd.MarkFlagRequired("doc-id")

	return cmd
}

func newDocTableShowCmd() *cobra.Command {
	var documentID, tableID string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "查看表格结构与内容",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			snapshot, err := fetchTableSnapshot(documentID, tableID)
			if err != nil {
				output.Errorf("获取表格失败: %v", err)
			}

			values := buildTableValues(snapshot)
			result := map[string]any{
				"table_id":       tableID,
				"rows":           snapshot.Rows,
				"cols":           snapshot.Cols,
				"header_row":     snapshot.HeaderRow,
				"header_column":  snapshot.HeaderColumn,
				"column_widths":  snapshot.ColumnWidths,
				"merge_info":     snapshot.MergeInfo,
				"cell_ids":       values.cellIDs,
				"values":         values.values,
				"table_block":    snapshot.Table,
				"descendant_cnt": len(snapshot.Blocks),
			}

			output.JSON(result)
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableWriteCmd() *cobra.Command {
	var documentID, tableID, data, dataFile string

	cmd := &cobra.Command{
		Use:   "write",
		Short: "重写整张表格的单元格文本",
		Long:  "按矩阵内容重写整张表格。每个单元格会清空后写入纯文本内容。",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			snapshot, err := fetchTableSnapshot(documentID, tableID)
			if err != nil {
				output.Errorf("获取表格失败: %v", err)
			}

			matrix, err := parseTableMatrixInput(data, dataFile, snapshot.Rows, snapshot.Cols)
			if err != nil {
				output.Errorf("解析表格数据失败: %v", err)
			}

			for row := 0; row < snapshot.Rows; row++ {
				for col := 0; col < snapshot.Cols; col++ {
					cellID, err := snapshot.cellID(row, col)
					if err != nil {
						output.Errorf("定位单元格失败: %v", err)
					}
					cellBlock := snapshot.Blocks[cellID]
					if err := replaceTableCellText(documentID, cellBlock, matrix[row][col]); err != nil {
						output.Errorf("写入单元格失败 (%d,%d): %v", row, col, err)
					}
				}
			}

			output.Success(fmt.Sprintf("表格写入成功 (%d 行 x %d 列)", snapshot.Rows, snapshot.Cols))
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().StringVar(&data, "data", "", "表格数据，支持 TSV/JSON 矩阵，- 表示从 stdin 读取")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "表格数据文件路径，支持 TSV 或 JSON 矩阵")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableSetCellCmd() *cobra.Command {
	var documentID, tableID, text string
	var row, col int

	cmd := &cobra.Command{
		Use:   "set-cell",
		Short: "更新单个单元格文本",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			snapshot, err := fetchTableSnapshot(documentID, tableID)
			if err != nil {
				output.Errorf("获取表格失败: %v", err)
			}

			cellID, err := snapshot.cellID(row, col)
			if err != nil {
				output.Errorf("定位单元格失败: %v", err)
			}

			if err := replaceTableCellText(documentID, snapshot.Blocks[cellID], text); err != nil {
				output.Errorf("更新单元格失败: %v", err)
			}

			output.Success(fmt.Sprintf("单元格更新成功 (%d,%d)", row, col))
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&row, "row", 0, "行索引，从 0 开始")
	cmd.Flags().IntVar(&col, "col", 0, "列索引，从 0 开始")
	cmd.Flags().StringVarP(&text, "text", "x", "", "新的单元格文本")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")
	cmd.MarkFlagRequired("row")
	cmd.MarkFlagRequired("col")
	cmd.MarkFlagRequired("text")

	return cmd
}

func newDocTableInsertRowCmd() *cobra.Command {
	var documentID, tableID, data, dataFile string
	var rowIndex int

	cmd := &cobra.Command{
		Use:   "insert-row",
		Short: "插入表格行",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				InsertTableRow(larkdocx.NewInsertTableRowRequestBuilder().RowIndex(rowIndex).Build()).
				Build()); err != nil {
				output.Errorf("插入行失败: %v", err)
			}

			if data == "" && dataFile == "" {
				output.Success("表格行插入成功")
				return
			}

			snapshot, err := fetchTableSnapshot(documentID, tableID)
			if err != nil {
				output.Errorf("插入行后获取表格失败: %v", err)
			}

			targetRow := rowIndex
			if targetRow < 0 {
				targetRow = snapshot.Rows - 1
			}

			values, err := parseVectorInput(data, dataFile, snapshot.Cols, "\t")
			if err != nil {
				output.Errorf("解析行数据失败: %v", err)
			}

			for col, value := range values {
				cellID, err := snapshot.cellID(targetRow, col)
				if err != nil {
					output.Errorf("定位新行单元格失败: %v", err)
				}
				if err := replaceTableCellText(documentID, snapshot.Blocks[cellID], value); err != nil {
					output.Errorf("写入新行单元格失败 (%d,%d): %v", targetRow, col, err)
				}
			}

			output.Success("表格行插入并写入成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&rowIndex, "row-index", -1, "插入行索引，-1 表示末尾")
	cmd.Flags().StringVar(&data, "data", "", "行数据，支持 TAB 分隔或 JSON 数组")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "行数据文件路径，支持 TAB 分隔或 JSON 数组")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableInsertColumnCmd() *cobra.Command {
	var documentID, tableID, data, dataFile string
	var columnIndex int

	cmd := &cobra.Command{
		Use:   "insert-column",
		Short: "插入表格列",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				InsertTableColumn(larkdocx.NewInsertTableColumnRequestBuilder().ColumnIndex(columnIndex).Build()).
				Build()); err != nil {
				output.Errorf("插入列失败: %v", err)
			}

			if data == "" && dataFile == "" {
				output.Success("表格列插入成功")
				return
			}

			snapshot, err := fetchTableSnapshot(documentID, tableID)
			if err != nil {
				output.Errorf("插入列后获取表格失败: %v", err)
			}

			targetCol := columnIndex
			if targetCol < 0 {
				targetCol = snapshot.Cols - 1
			}

			values, err := parseVectorInput(data, dataFile, snapshot.Rows, "\n")
			if err != nil {
				output.Errorf("解析列数据失败: %v", err)
			}

			for row, value := range values {
				cellID, err := snapshot.cellID(row, targetCol)
				if err != nil {
					output.Errorf("定位新列单元格失败: %v", err)
				}
				if err := replaceTableCellText(documentID, snapshot.Blocks[cellID], value); err != nil {
					output.Errorf("写入新列单元格失败 (%d,%d): %v", row, targetCol, err)
				}
			}

			output.Success("表格列插入并写入成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&columnIndex, "column-index", -1, "插入列索引，-1 表示末尾")
	cmd.Flags().StringVar(&data, "data", "", "列数据，默认按换行分隔，也支持 JSON 数组")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "列数据文件路径，默认按换行分隔，也支持 JSON 数组")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableDeleteRowsCmd() *cobra.Command {
	var documentID, tableID string
	var start, count int

	cmd := &cobra.Command{
		Use:   "delete-rows",
		Short: "删除表格行",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if count <= 0 {
				output.Errorf("count 必须大于 0")
			}

			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				DeleteTableRows(larkdocx.NewDeleteTableRowsRequestBuilder().
					RowStartIndex(start).
					RowEndIndex(start+count).
					Build()).
				Build()); err != nil {
				output.Errorf("删除表格行失败: %v", err)
			}

			output.Success("表格行删除成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&start, "start", 0, "起始行索引")
	cmd.Flags().IntVar(&count, "count", 1, "删除行数")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableDeleteColumnsCmd() *cobra.Command {
	var documentID, tableID string
	var start, count int

	cmd := &cobra.Command{
		Use:   "delete-columns",
		Short: "删除表格列",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if count <= 0 {
				output.Errorf("count 必须大于 0")
			}

			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				DeleteTableColumns(larkdocx.NewDeleteTableColumnsRequestBuilder().
					ColumnStartIndex(start).
					ColumnEndIndex(start+count).
					Build()).
				Build()); err != nil {
				output.Errorf("删除表格列失败: %v", err)
			}

			output.Success("表格列删除成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&start, "start", 0, "起始列索引")
	cmd.Flags().IntVar(&count, "count", 1, "删除列数")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableMergeCmd() *cobra.Command {
	var documentID, tableID string
	var rowStart, rowEnd, columnStart, columnEnd int

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "合并单元格",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				MergeTableCells(larkdocx.NewMergeTableCellsRequestBuilder().
					RowStartIndex(rowStart).
					RowEndIndex(rowEnd).
					ColumnStartIndex(columnStart).
					ColumnEndIndex(columnEnd).
					Build()).
				Build()); err != nil {
				output.Errorf("合并单元格失败: %v", err)
			}

			output.Success("单元格合并成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&rowStart, "row-start", 0, "行起始索引，左闭右开")
	cmd.Flags().IntVar(&rowEnd, "row-end", 1, "行结束索引，左闭右开")
	cmd.Flags().IntVar(&columnStart, "column-start", 0, "列起始索引，左闭右开")
	cmd.Flags().IntVar(&columnEnd, "column-end", 1, "列结束索引，左闭右开")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTableUnmergeCmd() *cobra.Command {
	var documentID, tableID string
	var row, col int

	cmd := &cobra.Command{
		Use:   "unmerge",
		Short: "取消合并单元格",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				UnmergeTableCells(larkdocx.NewUnmergeTableCellsRequestBuilder().
					RowIndex(row).
					ColumnIndex(col).
					Build()).
				Build()); err != nil {
				output.Errorf("取消合并单元格失败: %v", err)
			}

			output.Success("单元格取消合并成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&row, "row", 0, "行索引")
	cmd.Flags().IntVar(&col, "col", 0, "列索引")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func newDocTablePropsCmd() *cobra.Command {
	var documentID, tableID string
	var columnIndex, columnWidth int
	var headerRow, headerColumn bool

	cmd := &cobra.Command{
		Use:   "props",
		Short: "更新表格属性",
		Run: func(cmd *cobra.Command, args []string) {
			documentID = resolveDocumentID(documentID)
			builder := larkdocx.NewUpdateTablePropertyRequestBuilder()
			changed := false

			if cmd.Flags().Lookup("column-width").Changed {
				builder.ColumnWidth(columnWidth)
				builder.ColumnIndex(columnIndex)
				changed = true
			}
			if cmd.Flags().Lookup("header-row").Changed {
				builder.HeaderRow(headerRow)
				changed = true
			}
			if cmd.Flags().Lookup("header-column").Changed {
				builder.HeaderColumn(headerColumn)
				changed = true
			}

			if !changed {
				output.Errorf("至少需要指定一个属性变更")
			}

			if err := patchTableBlock(documentID, tableID, larkdocx.NewUpdateBlockRequestBuilder().
				UpdateTableProperty(builder.Build()).
				Build()); err != nil {
				output.Errorf("更新表格属性失败: %v", err)
			}

			output.Success("表格属性更新成功")
		},
	}

	cmd.Flags().StringVarP(&documentID, "doc-id", "d", "", "文档 ID 或 wiki 链接 (必填)")
	cmd.Flags().StringVarP(&tableID, "table-id", "t", "", "表格块 ID (必填)")
	cmd.Flags().IntVar(&columnIndex, "column-index", 0, "更新列宽时的列索引")
	cmd.Flags().IntVar(&columnWidth, "column-width", 0, "列宽，单位 px")
	cmd.Flags().BoolVar(&headerRow, "header-row", false, "是否启用标题行")
	cmd.Flags().BoolVar(&headerColumn, "header-column", false, "是否启用标题列")
	cmd.MarkFlagRequired("doc-id")
	cmd.MarkFlagRequired("table-id")

	return cmd
}

func buildTableDescendants(parentBlockID string, matrix [][]string, columnWidths []int, headerRow, headerColumn bool) (string, []*larkdocx.Block) {
	rows := len(matrix)
	cols := len(matrix[0])
	tableID := newTemporaryBlockID("table")
	cellIDs := make([]string, 0, rows*cols)
	descendants := make([]*larkdocx.Block, 0, 1+rows*cols*2)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cellID := newTemporaryBlockID(fmt.Sprintf("cell_%d_%d", row, col))
			cellIDs = append(cellIDs, cellID)
		}
	}

	tableBuilder := larkdocx.NewBlockBuilder().
		BlockId(tableID).
		ParentId(parentBlockID).
		Children(cellIDs).
		BlockType(docxBlockTypeTable).
		Table(larkdocx.NewTableBuilder().
			Cells(cellIDs).
			Property(larkdocx.NewTablePropertyBuilder().
				RowSize(rows).
				ColumnSize(cols).
				HeaderRow(headerRow).
				HeaderColumn(headerColumn).
				Build()).
			Build())
	if len(columnWidths) > 0 {
		tableBuilder.Table(larkdocx.NewTableBuilder().
			Cells(cellIDs).
			Property(larkdocx.NewTablePropertyBuilder().
				RowSize(rows).
				ColumnSize(cols).
				ColumnWidth(columnWidths).
				HeaderRow(headerRow).
				HeaderColumn(headerColumn).
				Build()).
			Build())
	}
	descendants = append(descendants, tableBuilder.Build())

	cellIndex := 0
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cellID := cellIDs[cellIndex]
			cellIndex++

			childIDs := []string{}
			cellBlock := larkdocx.NewBlockBuilder().
				BlockId(cellID).
				ParentId(tableID).
				Children(childIDs).
				BlockType(docxBlockTypeTableCell).
				TableCell(larkdocx.NewTableCellBuilder().Build())

			if matrix[row][col] != "" {
				textID := newTemporaryBlockID(fmt.Sprintf("cell_text_%d_%d", row, col))
				childIDs = append(childIDs, textID)
				cellBlock.Children(childIDs)
			}

			descendants = append(descendants, cellBlock.Build())

			if matrix[row][col] != "" {
				textID := childIDs[0]
				descendants = append(descendants, larkdocx.NewBlockBuilder().
					BlockId(textID).
					ParentId(cellID).
					BlockType(docxBlockTypeText).
					Text(newTextBody(matrix[row][col])).
					Build())
			}
		}
	}

	return tableID, descendants
}

func fetchTableSnapshot(documentID, tableID string) (*tableSnapshot, error) {
	pageToken := ""
	blocks := make([]*larkdocx.Block, 0)

	for {
		builder := larkdocx.NewGetDocumentBlockChildrenReqBuilder().
			DocumentId(documentID).
			BlockId(tableID).
			DocumentRevisionId(-1).
			PageSize(500).
			WithDescendants(true)
		if pageToken != "" {
			builder.PageToken(pageToken)
		}

		resp, err := larkClient.Docx.DocumentBlockChildren.Get(context.Background(), builder.Build())
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, fmt.Errorf("获取表格子孙块失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
		}
		if resp.Data != nil {
			blocks = append(blocks, resp.Data.Items...)
			if resp.Data.HasMore != nil && *resp.Data.HasMore && resp.Data.PageToken != nil {
				pageToken = *resp.Data.PageToken
				continue
			}
		}
		break
	}

	blockMap := make(map[string]*larkdocx.Block, len(blocks))
	for _, block := range blocks {
		if block != nil && block.BlockId != nil {
			blockMap[*block.BlockId] = block
		}
	}

	tableBlock, ok := blockMap[tableID]
	if !ok {
		resp, err := larkClient.Docx.DocumentBlock.Get(context.Background(), larkdocx.NewGetDocumentBlockReqBuilder().
			DocumentId(documentID).
			BlockId(tableID).
			Build())
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			return nil, fmt.Errorf("获取表格块失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
		}
		if resp.Data == nil || resp.Data.Block == nil {
			return nil, fmt.Errorf("表格块 %s 不存在", tableID)
		}
		tableBlock = resp.Data.Block
		blockMap[tableID] = tableBlock
	}

	if tableBlock.Table == nil {
		return nil, fmt.Errorf("块 %s 不是表格块", tableID)
	}

	rows := 0
	cols := 0
	headerRow := false
	headerColumn := false
	var widths []int
	var mergeInfo []*larkdocx.TableMergeInfo
	if tableBlock.Table.Property != nil {
		if tableBlock.Table.Property.RowSize != nil {
			rows = *tableBlock.Table.Property.RowSize
		}
		if tableBlock.Table.Property.ColumnSize != nil {
			cols = *tableBlock.Table.Property.ColumnSize
		}
		if tableBlock.Table.Property.HeaderRow != nil {
			headerRow = *tableBlock.Table.Property.HeaderRow
		}
		if tableBlock.Table.Property.HeaderColumn != nil {
			headerColumn = *tableBlock.Table.Property.HeaderColumn
		}
		widths = append(widths, tableBlock.Table.Property.ColumnWidth...)
		mergeInfo = tableBlock.Table.Property.MergeInfo
	}

	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("表格缺少合法的行列元数据")
	}

	return &tableSnapshot{
		Table:        tableBlock,
		Blocks:       blockMap,
		Rows:         rows,
		Cols:         cols,
		CellIDs:      append([]string{}, tableBlock.Table.Cells...),
		HeaderRow:    headerRow,
		HeaderColumn: headerColumn,
		ColumnWidths: widths,
		MergeInfo:    mergeInfo,
	}, nil
}

func (s *tableSnapshot) cellID(row, col int) (string, error) {
	if row < 0 || row >= s.Rows {
		return "", fmt.Errorf("row=%d 超出范围 [0,%d)", row, s.Rows)
	}
	if col < 0 || col >= s.Cols {
		return "", fmt.Errorf("col=%d 超出范围 [0,%d)", col, s.Cols)
	}

	index := row*s.Cols + col
	if index >= len(s.CellIDs) {
		return "", fmt.Errorf("表格 cells 数量不足: 期望至少 %d，实际 %d", index+1, len(s.CellIDs))
	}
	return s.CellIDs[index], nil
}

func patchTableBlock(documentID, tableID string, updateReq *larkdocx.UpdateBlockRequest) error {
	resp, err := larkClient.Docx.DocumentBlock.Patch(context.Background(), larkdocx.NewPatchDocumentBlockReqBuilder().
		DocumentId(documentID).
		BlockId(tableID).
		DocumentRevisionId(-1).
		UpdateBlockRequest(updateReq).
		Build())
	if err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("[%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
	}
	return nil
}

func replaceTableCellText(documentID string, cellBlock *larkdocx.Block, text string) error {
	if cellBlock == nil || cellBlock.BlockId == nil {
		return fmt.Errorf("单元格块不存在")
	}

	if len(cellBlock.Children) > 0 {
		resp, err := larkClient.Docx.DocumentBlockChildren.BatchDelete(context.Background(), larkdocx.NewBatchDeleteDocumentBlockChildrenReqBuilder().
			DocumentId(documentID).
			BlockId(*cellBlock.BlockId).
			DocumentRevisionId(-1).
			Body(larkdocx.NewBatchDeleteDocumentBlockChildrenReqBodyBuilder().
				StartIndex(0).
				EndIndex(len(cellBlock.Children)).
				Build()).
			Build())
		if err != nil {
			return err
		}
		if !resp.Success() {
			return fmt.Errorf("清空单元格失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
		}
	}

	if text == "" {
		return nil
	}

	resp, err := larkClient.Docx.DocumentBlockChildren.Create(context.Background(), larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
		DocumentId(documentID).
		BlockId(*cellBlock.BlockId).
		DocumentRevisionId(-1).
		Body(larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
			Children([]*larkdocx.Block{
				larkdocx.NewBlockBuilder().
					BlockType(docxBlockTypeText).
					Text(newTextBody(text)).
					Build(),
			}).
			Index(0).
			Build()).
		Build())
	if err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("写入单元格失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
	}

	return nil
}

func parseTableMatrixInput(data, dataFile string, rows, cols int) ([][]string, error) {
	raw, err := readOptionalInput(data, dataFile)
	if err != nil {
		return nil, err
	}

	matrix := [][]string{}
	if strings.TrimSpace(raw) != "" {
		matrix, err = parseMatrix(raw)
		if err != nil {
			return nil, err
		}
	}

	return normalizeMatrix(matrix, rows, cols)
}

func parseVectorInput(data, dataFile string, size int, separator string) ([]string, error) {
	raw, err := readOptionalInput(data, dataFile)
	if err != nil {
		return nil, err
	}

	values := []string{}
	trimmed := strings.TrimSpace(raw)
	if trimmed != "" {
		if strings.HasPrefix(trimmed, "[") {
			if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
				return nil, fmt.Errorf("解析 JSON 数组失败: %w", err)
			}
		} else {
			parts := strings.Split(strings.TrimRight(raw, "\r\n"), separator)
			values = make([]string, 0, len(parts))
			for _, part := range parts {
				values = append(values, strings.TrimSuffix(part, "\r"))
			}
		}
	}

	if len(values) > size {
		return nil, fmt.Errorf("数据长度 %d 超过目标长度 %d", len(values), size)
	}

	result := make([]string, size)
	copy(result, values)
	return result, nil
}

func parseMatrix(raw string) ([][]string, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "[") {
		var matrix [][]string
		if err := json.Unmarshal([]byte(trimmed), &matrix); err != nil {
			return nil, fmt.Errorf("解析 JSON 矩阵失败: %w", err)
		}
		return matrix, nil
	}

	raw = strings.TrimRight(raw, "\r\n")
	if raw == "" {
		return [][]string{}, nil
	}

	lines := strings.Split(raw, "\n")
	matrix := make([][]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		matrix = append(matrix, strings.Split(line, "\t"))
	}
	return matrix, nil
}

func normalizeMatrix(matrix [][]string, rows, cols int) ([][]string, error) {
	if rows < 0 || cols < 0 {
		return nil, fmt.Errorf("rows/cols 不能为负数")
	}

	if rows == 0 {
		rows = len(matrix)
	}
	if cols == 0 {
		for _, row := range matrix {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}

	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("必须提供有效的 rows/cols 或非空 data")
	}

	if len(matrix) > rows {
		return nil, fmt.Errorf("数据行数 %d 超过指定行数 %d", len(matrix), rows)
	}

	result := make([][]string, rows)
	for row := 0; row < rows; row++ {
		result[row] = make([]string, cols)
		if row >= len(matrix) {
			continue
		}
		if len(matrix[row]) > cols {
			return nil, fmt.Errorf("第 %d 行列数 %d 超过指定列数 %d", row, len(matrix[row]), cols)
		}
		copy(result[row], matrix[row])
	}

	return result, nil
}

func readOptionalInput(data, dataFile string) (string, error) {
	if data != "" && dataFile != "" {
		return "", fmt.Errorf("data 和 data-file 不能同时指定")
	}
	if dataFile == "" {
		return readStdinOrValue(data)
	}

	content, err := os.ReadFile(dataFile)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func parseColumnWidths(raw string, cols int) ([]int, error) {
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	if len(parts) != cols {
		return nil, fmt.Errorf("列宽数量 %d 与列数 %d 不一致", len(parts), cols)
	}

	widths := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("非法列宽 %q", part)
		}
		widths = append(widths, value)
	}
	return widths, nil
}

func newTemporaryBlockID(prefix string) string {
	replacer := strings.NewReplacer(" ", "_", "-", "_")
	return fmt.Sprintf("tmp_%s", replacer.Replace(prefix))
}

func newTextBody(content string) *larkdocx.Text {
	return larkdocx.NewTextBuilder().
		Elements([]*larkdocx.TextElement{
			larkdocx.NewTextElementBuilder().
				TextRun(larkdocx.NewTextRunBuilder().Content(content).Build()).
				Build(),
		}).
		Build()
}

type tableValues struct {
	values  [][]string
	cellIDs [][]string
}

func buildTableValues(snapshot *tableSnapshot) tableValues {
	values := make([][]string, snapshot.Rows)
	cellIDs := make([][]string, snapshot.Rows)

	for row := 0; row < snapshot.Rows; row++ {
		values[row] = make([]string, snapshot.Cols)
		cellIDs[row] = make([]string, snapshot.Cols)
		for col := 0; col < snapshot.Cols; col++ {
			cellID, err := snapshot.cellID(row, col)
			if err != nil {
				continue
			}
			cellIDs[row][col] = cellID
			values[row][col] = tableCellText(snapshot.Blocks, cellID)
		}
	}

	return tableValues{values: values, cellIDs: cellIDs}
}

func tableCellText(blocks map[string]*larkdocx.Block, cellID string) string {
	cellBlock := blocks[cellID]
	if cellBlock == nil || len(cellBlock.Children) == 0 {
		return ""
	}

	parts := make([]string, 0, len(cellBlock.Children))
	for _, childID := range cellBlock.Children {
		if text := blockPlainText(blocks[childID]); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func blockPlainText(block *larkdocx.Block) string {
	if block == nil {
		return ""
	}

	switch {
	case block.Text != nil:
		return textElementsPlain(block.Text.Elements)
	case block.Heading1 != nil:
		return textElementsPlain(block.Heading1.Elements)
	case block.Heading2 != nil:
		return textElementsPlain(block.Heading2.Elements)
	case block.Heading3 != nil:
		return textElementsPlain(block.Heading3.Elements)
	case block.Heading4 != nil:
		return textElementsPlain(block.Heading4.Elements)
	case block.Heading5 != nil:
		return textElementsPlain(block.Heading5.Elements)
	case block.Heading6 != nil:
		return textElementsPlain(block.Heading6.Elements)
	case block.Heading7 != nil:
		return textElementsPlain(block.Heading7.Elements)
	case block.Heading8 != nil:
		return textElementsPlain(block.Heading8.Elements)
	case block.Heading9 != nil:
		return textElementsPlain(block.Heading9.Elements)
	case block.Bullet != nil:
		return textElementsPlain(block.Bullet.Elements)
	case block.Ordered != nil:
		return textElementsPlain(block.Ordered.Elements)
	case block.Code != nil:
		return textElementsPlain(block.Code.Elements)
	case block.Quote != nil:
		return textElementsPlain(block.Quote.Elements)
	case block.Todo != nil:
		return textElementsPlain(block.Todo.Elements)
	default:
		return ""
	}
}

func textElementsPlain(elements []*larkdocx.TextElement) string {
	if len(elements) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, element := range elements {
		if element == nil {
			continue
		}
		if element.TextRun != nil && element.TextRun.Content != nil {
			builder.WriteString(*element.TextRun.Content)
		}
	}
	return builder.String()
}
