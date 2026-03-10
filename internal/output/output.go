package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// JSON 以 JSON 格式输出数据
func JSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 序列化失败: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// Table 以表格形式输出数据
func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)

	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, col)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}

// Success 输出成功信息
func Success(msg string) {
	fmt.Printf("✓ %s\n", msg)
}

// Error 输出错误信息并退出
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	os.Exit(1)
}

// Errorf 格式化输出错误信息并退出
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}
