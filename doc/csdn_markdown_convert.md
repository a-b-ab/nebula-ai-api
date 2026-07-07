# Markdown 转 CSDN HTML

## 普通单元测试

```bash
GOCACHE=/private/tmp/nebula-go-build-cache go test ./internal/service -run TestMarkdownServiceConvertToCSDN -v
```

该测试会读取 `internal/service/testdata/csdn_sample.md`，验证标题、列表、引用、表格、代码块和 CSDN 包装容器是否正常输出。

## 转换外部 Markdown 文件

```bash
CSDN_MD_FILE=/path/to/article.md \
CSDN_TITLE="文章标题" \
CSDN_HTML_OUT=/private/tmp/article.csdn.html \
GOCACHE=/private/tmp/nebula-go-build-cache \
go test ./internal/service -run TestMarkdownServiceConvertExternalMarkdownFileToCSDN -v
```

生成后的 `/private/tmp/article.csdn.html` 可以复制到 CSDN 富文本/HTML 内容区域使用。
