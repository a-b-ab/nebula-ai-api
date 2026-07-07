# Markdown 转 CSDN 示例

这是一篇用于验证 Markdown 转 CSDN HTML 的样例文章。

## 功能点

- 保留标题、列表、引用、表格和代码块
- 支持 GFM 表格
- 保留代码语言标记，方便 CSDN 识别高亮

> 这段引用用于确认 blockquote 能被正确输出。

| 语法 | 期望效果 |
| --- | --- |
| `# 标题` | 输出 h1/h2 等标题 |
| 代码块 | 输出带 language 的 code 标签 |

```go
package main

import "fmt"

func main() {
	fmt.Println("hello csdn")
}
```
