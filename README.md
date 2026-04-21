# claude-fish

一个伪装成 CLI 编程工具的终端小说阅读器。

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

在终端里看小说，但看起来就像在写代码。支持一键切换到「老板模式」，瞬间显示一段带语法高亮的代码流式输出，完美通过领导检查。

## 功能特点

- **三种 CLI 视觉风格** — Claude Code / Codex CLI / opencode，阅读时随时按键切换
- **老板键保护** — 按 `Tab` 瞬间从小说切换到代码编辑界面，代码逐字流式输出，带语法高亮
- **多种小说格式** — 支持 TXT、Markdown、EPUB，根据文件扩展名自动识别
- **CJK 友好** — 中日韩字符按双宽度计算，不会出现截断乱码
- **单二进制分发** — 编译后一个文件，零依赖，下载即用

## 安装

### 从源码编译

需要 Go 1.24 或更高版本。

```bash
git clone https://github.com/tony/claude-fish.git
cd claude-fish
go build -o claude-fish .
```

### 直接下载

前往 [Releases](https://github.com/tony/claude-fish/releases) 页面下载对应平台的二进制文件。

## 使用方法

### 基本用法

```bash
# 阅读一本小说
claude-fish novel.txt

# 阅读小说，并指定一个代码文件作为老板键保护
claude-fish novel.epub -c main.go

# 使用 Codex CLI 风格
claude-fish novel.txt -t codex

# 使用 opencode 风格
claude-fish novel.md -t opencode

# 自定义流式输出速度（毫秒/字符）
claude-fish novel.txt -c handler.go --speed 15
```

### 命令行参数

| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--code` | `-c` | 代码文件路径，用于老板模式 | 无 |
| `--theme` | `-t` | 视觉主题：`claude`、`codex`、`opencode` | `claude` |
| `--speed` | | 流式输出速度（毫秒/字符） | `25` |
| `--help` | `-h` | 显示帮助信息 | |

### 快捷键

| 按键 | 欢迎页 | 阅读模式 | 老板模式 |
|------|--------|----------|----------|
| `Space` / `Enter` | 开始阅读 | 下一页 | — |
| `B` / `←` / `H` | — | 上一页 | — |
| `S` | — | 切换主题风格 | — |
| `Tab` | — | 进入老板模式 | 返回阅读 |
| `Esc` | — | — | 返回阅读 |
| `Q` / `Ctrl+C` | 退出 | 退出 | 退出 |

## 视觉风格

### Claude Code（默认）

紫色 (#7c3aed) 边框和强调色，橙色 (#f97316) ASCII Logo，对话气泡风格的内容展示，模拟 Claude Code 的完整界面布局。

```
╭─── claude-fish v1.0.0 ──────────────────────────────────────╮
│                  Welcome!                  │ Loaded 三体.txt │
│                      ▐▛███▜▌              │ 42 chapters     │
│                     ▝▜█████▛▘              │ Press Space     │
│                       ▘▘ ▝▝               │ Press Tab       │
│   claude · v1.0.0                         │                 │
╰────────────────────────────────────────────────────────────╯
────────────────────────────────────────────────────────────
❯
```

### Codex CLI

绿色 (#10a37f) 强调色，极简风格，带有方块进度条显示阅读进度。

```
codex v1.0.0
──────────────────────────────────────────────────────────
Loaded: 三体.txt (42 chapters)

Press Space to start reading
```

### opencode

Catppuccin Mocha 配色方案，顶部 Tab 栏设计，柔和的蓝灰色调。

```
Welcome   Files   Config
──────
 ╭──────────────────────────────────────────────────╮
 │ opencode                                          │
 │                                                   │
 │ Loaded: 三体.txt (42 chapters)                    │
 │ Press Space to start                               │
 ╰──────────────────────────────────────────────────╯
```

## 老板模式

老板模式是 claude-fish 的核心功能。使用 `-c` 参数指定一个代码文件，当需要时按 `Tab` 键即可瞬间切换到代码编辑界面。

### 工作原理

1. **瞬间切换** — 按 `Tab` 立即从小说阅读界面切换到代码编辑界面，零延迟
2. **流式输出** — 代码逐字符流式显示，带有随机速度抖动（±40%），模拟真实的 AI 编程工具输出
3. **语法高亮** — 使用 Chroma 库根据文件扩展名自动识别语言并高亮显示
4. **AI 前言** — 流式输出代码前会先显示一段「AI 思考」提示文本
5. **循环播放** — 代码输出完毕后自动从头循环
6. **断点续传** — 按 `Tab` 切回阅读模式时暂停，再次进入时从上次位置继续

### 示例

```bash
# 准备一个看起来像在写的代码文件
claude-fish 三体.txt -c src/api/handler.go

# 阅读时按 Tab → 瞬间显示如下界面：
╭────────────────────────────────────────────────────────────╮
│ ● Editing handler.go │ claude-fish                        │
╰────────────────────────────────────────────────────────────╯
✦ Let me implement the changes in handler.go:
┌─ handler.go
package api

import (
    "net/http"
    "encoding/json"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)▌
Writing to handler.go...
```

## 支持的小说格式

| 格式 | 扩展名 | 说明 |
|------|--------|------|
| 纯文本 | `.txt` | 按空行分章节，以「第」或「Chapter」开头的行识别为章节标题 |
| Markdown | `.md` `.markdown` | 按 `#` / `##` 标题分章节 |
| EPUB | `.epub` | 标准 EPUB 格式，自动解析 XHTML 内容和章节结构 |

## 项目结构

```
claude-fish/
├── main.go                    # 入口
├── cmd/
│   └── root.go                # Cobra 命令定义和参数解析
├── internal/
│   ├── app.go                 # Bubble Tea 主模型，三态状态机
│   ├── pager.go               # 分页引擎，CJK 双宽度感知
│   ├── boss.go                # 老板模式状态管理
│   ├── streamer.go            # 代码流式输出引擎
│   ├── highlight.go           # Chroma 语法高亮
│   ├── theme/
│   │   ├── theme.go           # Theme 接口定义
│   │   ├── claudecode.go      # Claude Code 风格
│   │   ├── codex.go           # Codex CLI 风格
│   │   └── opencode.go        # opencode 风格
│   └── reader/
│       ├── reader.go          # Reader 接口定义
│       ├── txt.go             # TXT 解析器
│       ├── markdown.go        # Markdown 解析器
│       └── epub.go            # EPUB 解析器
└── testdata/
    ├── sample.txt
    └── sample.md
```

## 技术栈

| 库 | 用途 |
|----|------|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI 框架 |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | 终端样式渲染 |
| [Bubbles](https://github.com/charmbracelet/bubbles) | 预制组件 |
| [Cobra](https://github.com/spf13/cobra) | CLI 命令和参数解析 |
| [Chroma](https://github.com/alecthomas/chroma) | 代码语法高亮 |

## 开发

```bash
# 克隆仓库
git clone https://github.com/tony/claude-fish.git
cd claude-fish

# 安装依赖
go mod download

# 运行测试
go test ./...

# 编译
go build -o claude-fish .

# 运行
./claude-fish testdata/sample.txt -c main.go
```

## License

MIT
