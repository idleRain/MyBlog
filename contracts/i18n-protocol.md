# i18n 内容语言协商契约

> 本文是多语言内容协商的唯一契约锚点。改 wire 格式（语言协商语义、翻译字段形状）必须先改本文。
> UI 静态文案不走本契约，前台界面文案由 `apps/web` paraglide 词表体系承接。

## 状态

✅ 已落地：语言解析中间件、响应本地化与 `Content-Language` 标注、翻译写入与搜索合并均已实现，
覆盖文章、分类、标签与字典四类内容实体。实现见 `internal/middleware/language.go`、
`internal/domain/language.go`、`internal/service/localization.go`、`internal/handler/localization.go`。

## 语言协商

| 事实 | 约定 |
|---|---|
| 请求头 | `Accept-Language`，HTTP 标准语义 |
| 取值 | 具体语言标签（`zh`、`en`，地区变体如 `zh-CN` 折叠为主子标签）或通配符 `*` |
| 白名单权威 | `server/configs/config.yaml` 的 `i18n` 节（`default_language` + `supported_languages`） |
| 缺省行为 | 请求头为空、白名单外标签均回退 `default_language`（当前为 `zh`） |
| 全量包语义 | `Accept-Language: *` 表示响应携带该实体的全部翻译行，供管理端编辑使用 |
| 解析规则 | 多标签按出现顺序取首个白名单命中，未实现质量值加权匹配（`;q=` 参数在归一化时剥离，不参与排序） |
| 上下文传递 | 中间件解析后写入请求上下文，键为 `domain.LanguageContextKey`，handler 层读取 |
| 中间件挂载 | 语言中间件为全局中间件，业务 handler 始终能读到明确的语言标识，无需各自解析请求头 |

## 响应约定

| 事实 | 约定 |
|---|---|
| 实际语言标注 | 单资源读接口成功响应附带 `Content-Language` 头，值为实际输出语言；发生回退时标注回退后语言 |
| 集合响应标注 | 列表类响应的 `Content-Language` 为请求语言，集合内未翻译项按字段回退默认语言 |
| 全量包标注 | 全量包响应的 `Content-Language` 为 `*` |
| 字段形状 | 本地化输出保持既有 json 字段名不变，仅替换字段值为目标语言内容 |
| 翻译缺失 | 目标语言缺少翻译的字段回退默认语言内容，响应不携带空字段 |
| 翻译行字段 | `translations` 仅在 `Accept-Language: *` 下输出；具体语言请求输出前清空该字段，公共响应不携带全量翻译包 |
| 翻译状态标记 | `translationLocales` 为已填写有效翻译内容的语言列表，按字典序输出；当前仅文章实体具备该字段，具体语言与全量包请求均携带 |

### 公共读接口响应示例（`Accept-Language: en`，翻译齐全）

```jsonc
{
  "title": "Hello World",
  "summary": "English summary",
  "content": "English content"
}
```

### 管理端读接口响应示例（`Accept-Language: *`）

```jsonc
{
  "title": "你好世界",                       // 主列恒为默认语言内容
  "translations": [                         // 全量翻译行数组，仅 Accept-Language: * 下输出
    { "locale": "en", "title": "Hello World", "summary": "...", "content": "..." }
  ],
  "translationLocales": ["en"]              // 文章专属的轻量翻译状态标记，不含翻译内容
}
```

## 翻译写入契约

| 事实 | 约定 |
|---|---|
| 请求字段 | 创建与更新请求体新增可选 `i18n` 对象，键为语言标识，值为该语言字段对象；文章、分类、标签、字典类型与字典项的创建/更新请求均支持 |
| 主语言规则 | 默认语言内容始终走主字段，`i18n` 中出现默认语言键视为非法输入 |
| 更新语义 | `i18n` 中提供的字段提供即更新，未提供的字段保持原值，与主字段更新语义对齐；写入以既有翻译行为底合并补丁字段 |
| 写入时点 | 文章翻译行随 `CreateWithRelations` / `UpdateWithRelations` 的主事务写入（与分类标签关联同步同事务）；分类、标签、字典的翻译行在实体本体写入之后单独提交，经 `UpsertTranslations` / `UpsertTypeTranslations` / `UpsertItemTranslations` 单事务写入，命中复合唯一索引时整行更新，翻译写入失败不回滚实体本体 |
| 校验规则 | 语言键必须在白名单内；字段长度校验与主字段同规则；校验失败映射 400 |
| 派生字段 | 翻译 `content` 保存时按主字段同管线渲染 `contentHtml` 并统计 `wordCount` |

### 写入请求示例

```jsonc
{
  "id": 1,
  "title": "你好世界",
  "i18n": {
    "en": { "title": "Hello World", "summary": "English summary" }
  }
}
```

## 范围

| 实体 | 多语言字段 |
|---|---|
| 文章 | `title` / `summary` / `content` / `contentHtml` / `wordCount` / `seoTitle` / `seoDescription` / `seoKeywords` |
| 分类 | `name` / `description` / `seoTitle` / `seoDescription` |
| 标签 | `name` / `description` |
| 字典类型 | `name` / `description` |
| 字典项 | `label` / `description` |

搜索合并：文章关键词搜索同时匹配主表与 `article_translations`（标题、正文、摘要的全文索引），
其余实体无关键词搜索接口，不涉及翻译表匹配。

友链、站点设置公开文本项、通知文案、评论与用户资料不在本契约范围；文章修订快照暂不含翻译内容。
