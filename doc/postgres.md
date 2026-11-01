# 文章表
content字段用text类型存
PostgreSQL 的 TEXT 性能特点：
1. TOAST机制
- PostgreSQL 会自动压缩和存储大字段
- 如果字段超过约 2KB，会自动存储到单独的 TOAST 表中
- 主表只保留指针，不影响查询性能
2. TEXT vs VARCHAR 性能几乎相同
- 在 PostgreSQL 内部，两者实现方式基本一样
- 对于小内容（< 2KB），都存在主表
- 对于大内容（> 2KB），都用 TOAST