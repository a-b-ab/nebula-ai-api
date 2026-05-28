# 星云AI（Nebula AI）

一个基于 Go 构建的 AI 内容工作流平台（AI Content Workflow Platform），聚焦 AIGC 内容生产与自动化发布场景。

平台集成：

* 文生文（Text-to-Text）
* 文生图（Text-to-Image）
* 图生文（Image-to-Text）
* 图生图（Image-to-Image）
* AI 内容优化
* Markdown 内容管理
* 多平台自动发布

通过统一 AI Gateway 与异步任务系统，实现 AI 能力工程化、平台化与工作流化。

---

# 项目目标

## 核心目标

* 构建可扩展的 AI 内容生产平台
* 实践 Go 后端工程化与 AI Workflow 架构
* 打造完整 AIGC 内容生产体系
* 实现 AI 能力统一接入与任务调度
* 沉淀 AI 应用层工程经验

---

## 技术目标

* AI Provider 抽象与模型统一管理
* 异步任务调度与失败重试
* Markdown 内容工程化
* 浏览器自动化发布
* AI Workflow Pipeline
* 多平台 Adapter 架构

---

# 核心功能

## 1. AI 内容工作流（核心）

支持：

* AI 文章优化
* AI SEO 标题生成
* AI 摘要生成
* AI 标签生成
* 平台风格改写
* Markdown 内容增强

典型流程：

```text
Markdown
    ↓
AI优化
    ↓
SEO生成
    ↓
平台改写
    ↓
自动发布
```

---

## 2. 多模态 AI 能力

### 文生文（Text-to-Text）

支持：

* 技术文章优化
* 内容改写
* 摘要生成
* Prompt Workflow

---

### 文生图（Text-to-Image）

支持：

* AI 自动封面生成
* 技术风格插图
* 架构图生成

---

### 图生文（Image-to-Text）

支持：

* 架构图解析
* OCR 内容提取
* 图片内容描述

---

### 图生图（Image-to-Image）

支持：

* AI 图片增强
* 风格转换
* 封面优化

---

# 3. 自动发布系统（核心）

支持：

* 掘金
* 博客园（规划中）
* CSDN（规划中）
* 知乎（规划中）

基于：

* Playwright
* Browser Automation
* Adapter Pattern

实现多平台自动发布。

---

# 4. AI Gateway

统一管理：

* OpenAI
* DeepSeek
* Claude
* Gemini
* Qwen
* Stable Diffusion（规划中）

通过 Provider 模式实现统一 AI 能力抽象。

---

# 5. 异步任务系统

基于 Redis Queue 实现：

* AI 推理任务
* 发布任务
* Retry 重试
* 幂等控制
* 状态流转

---

# 6. Markdown 内容管理

支持：

* Markdown 编辑
* 实时预览
* AI 增强
* 内容版本管理

---

# 技术栈

## 后端

```text
Go
Gin
GORM
PostgreSQL
Redis
```

---

## AI 能力层

```text
OpenAI Compatible API
DeepSeek
Claude
Gemini
Qwen
```

---

## 自动化

```text
Node.js
Playwright
```

---

## 前端

```text
Next.js
React
TypeScript
TailwindCSS
shadcn/ui
```

---

# 系统架构

```text
┌────────────────────┐
│   React Frontend   │
└─────────┬──────────┘
          │
┌─────────▼──────────┐
│     Go Backend     │
├────────────────────┤
│ Article Module     │
│ AI Workflow Module │
│ Publish Module     │
│ Task Module        │
└─────────┬──────────┘
          │
 ┌────────▼────────┐
 │  PostgreSQL     │
 └─────────────────┘

 ┌─────────────────┐
 │   Redis Queue   │
 └────────┬────────┘
          │
 ┌────────▼────────┐
 │   Go Worker     │
 └────────┬────────┘
          │
 ┌────────▼────────┐
 │ AI Gateway      │
 └─────────────────┘

 ┌─────────────────┐
 │ Playwright      │
 │ Publisher       │
 └─────────────────┘
```

---

# 核心设计

## 1. Provider Pattern

统一 AI 模型接口：

```go
type AIProvider interface {
    Chat(ctx context.Context, req *ChatReq) error
}
```

---

## 2. Adapter Pattern

统一发布平台接口：

```go
type Publisher interface {
    Publish(ctx context.Context, article *Article) error
}
```

---

## 3. Workflow Pipeline

通过 Pipeline 实现 AI 工作流：

```text
Content
    ↓
Optimize
    ↓
SEO
    ↓
Rewrite
    ↓
Publish
```

---

# MVP 功能

当前 MVP 聚焦：

* Markdown 写作
* AI 文章优化
* 掘金自动发布
* Redis 异步任务
* AI Provider 接入

验证：

* AI Workflow
* Browser Automation
* 内容工程化
* 异步任务调度

---

# 后续规划

## V2

* 多平台发布
* AI 封面生成
* AI 摘要生成
* 定时发布

---

## V3

* 多 Agent Workflow
* 图生文
* Mermaid 自动生成
* 内容分析

---

## V4

* AI 内容运营系统
* 数据分析
* Workflow 编排
* 插件系统

---

# 项目价值

该项目不仅是 AI 应用实践项目，同时也是：

* Go 工程化实践
* AI Workflow 架构实践
* AIGC 内容平台实践
* 自动化工作流实践
* AI 工程能力沉淀项目

适用于：

* AI 应用开发
* SaaS 平台架构
* 内容自动化场景
* AI Workflow 系统设计
* Go 后端工程化实践
