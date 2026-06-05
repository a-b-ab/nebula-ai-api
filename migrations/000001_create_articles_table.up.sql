CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,

    -- AI 生成内容
    summary VARCHAR(1000),
    tags VARCHAR(1000),

    -- AI 优化状态
    ai_optimized BOOLEAN DEFAULT FALSE,

    -- 发布平台（逗号分隔）
    publish_platform VARCHAR(200),

    -- 文章状态（0: draft, 1: optimized, 2: published）
    status SMALLINT DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_ai_optimized ON articles(ai_optimized);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);
