# claude-fish — Terminal Novel Reader Design

A terminal-based novel reading tool that mimics the visual style of CLI programming tools (Claude Code, Codex CLI, opencode), with a "boss key" feature that instantly switches to a realistic code streaming view.

## Architecture

```
claude-fish/
├── main.go                  # Entry point, CLI argument parsing
├── internal/
│   ├── app.go               # Bubble Tea main model, state machine
│   ├── theme/
│   │   ├── theme.go         # Theme interface definition
│   │   ├── claudecode.go    # Claude Code style (purple + orange logo)
│   │   ├── codex.go         # Codex CLI style (green + progress bar)
│   │   └── opencode.go      # opencode style (Catppuccin palette + tab bar)
│   ├── reader/
│   │   ├── reader.go        # Reader interface
│   │   ├── epub.go          # EPUB parser
│   │   ├── txt.go           # TXT parser
│   │   └── markdown.go      # Markdown parser
│   ├── pager.go             # Pagination engine (splits by terminal height)
│   ├── boss.go              # Boss mode state management
│   └── streamer.go          # Code streaming engine (char-by-char + delay)
├── cmd/
│   └── root.go              # cobra command definitions
└── go.mod
```

**Core design principles:**

- **State machine driven**: Three states — `Welcome` → `Reading` → `BossMode`. Tab switches instantly between Reading and BossMode.
- **Theme interface**: Each CLI style implements a unified `Theme` interface. Switching styles swaps only the Theme implementation.
- **Reader interface**: Unified pagination interface. EPUB/TXT/Markdown each implement parsing, outputting plain text + chapter structure to the Pager.

## Theme System

```go
type Theme interface {
    Name() string
    RenderWelcome(info WelcomeInfo) string    // Welcome page
    RenderPage(page PageInfo) string          // Reading page (novel content)
    RenderCode(code CodeInfo) string          // Boss mode (streaming code)
    RenderStatusBar(status StatusInfo) string // Bottom status bar / key hints
    AccentColor() lipgloss.Color              // Theme accent color
}
```

Each Theme implementation determines: border style, color scheme, Logo/ASCII art, progress display style.

### Theme Variants

**Claude Code** (default):
- Purple (#7c3aed) border and accent, orange (#f97316) ASCII art logo
- Rounded box header with file/model info
- Chat bubble style content (left purple border + dark purple background)
- Tool call indicators (green checkmarks)

**Codex CLI**:
- Green (#10a37f) accent on dark background (#0d1117)
- Plain `>` prompt style
- Block progress bar for page position
- Minimal status line

**opencode**:
- Catppuccin Mocha palette (#89b4fa, #cdd6f4, #1e1e2e)
- Tab bar at top (Chat/Files/Diff)
- Clean bordered content blocks

## State Machine

```
Welcome ──Enter──→ Reading ←──Tab──→ BossMode
                     ↑                  │
                     └──── Tab ──────────┘

Reading 中按 S 切换 Theme（不影响当前页，即时刷新）
```

### Key Bindings

| Key | Welcome | Reading | BossMode |
|-----|---------|---------|----------|
| Space | Start reading | Next page | — |
| B | — | Previous page | — |
| S | — | Switch theme | — |
| Tab | — | → BossMode | → Reading |
| Q | Quit | Quit | Quit |
| Esc | — | — | → Reading |

## Reader & Pagination

```go
type Reader interface {
    Load(path string) error
    Chapters() []Chapter
    ReadPage(chapter, page int) string
    TotalPages(chapter int) int
}

type Chapter struct {
    Title string
    Index int
}
```

### Format Handling

- **TXT**: Split by blank lines (double newline `\n\n`). Use "Chapter N" when no explicit title.
- **Markdown**: Split by `#`/`##` headings. Preserve chapter titles.
- **EPUB**: Unzip and parse XHTML via `go-epub`. Extract body text, follow EPUB chapter structure.

### Pagination Engine

```go
type Pager struct {
    width  int    // Terminal width
    height int    // Available height (minus borders and status bar)
    theme  Theme  // Current theme (affects border line count)
}

func (p *Pager) Paginate(text string) []string
```

- Re-paginate on terminal resize
- CJK characters count as double-width to prevent line-break corruption
- Leave 1 blank line at bottom of each page for visual breathing room

## Boss Mode & Code Streaming

```go
type BossMode struct {
    code      string         // Code file content
    codeFile  string         // Filename for title bar display
    theme     Theme
    displayed int            // Characters already output
    ticker    *time.Ticker
}
```

### Streaming Engine

- Load user-specified code file entirely into memory
- Begin streaming immediately on entering BossMode
- Speed: 15-30ms per character with random jitter for realistic typing feel
- Insert "tool call" hints every 3-5 lines (e.g., `✔ Reading file...`, `Writing to src/api/handler.go...`)
- Stop and blink cursor when code is fully output

### Realism Details

1. Title bar changes from "Reading 三体.txt" to "Editing src/api/handler.go"
2. Code has syntax highlighting via Chroma (auto-detect language from file extension)
3. Before streaming code, output an "AI thinking" preamble (e.g., `✦ Let me implement the API handler with proper error handling:`), also char-by-char
4. On Tab back to Reading, Streamer pauses. Re-entering BossMode resumes from paused position. When code finishes, loop from beginning.

## CLI Usage

```bash
# Basic usage
claude-fish novel.epub

# With boss key protection file
claude-fish novel.epub -c main.go

# Specify initial theme
claude-fish novel.txt -t codex

# Set streaming speed
claude-fish novel.md -c handler.go --speed 20
```

### Arguments

| Flag | Description | Default |
|------|-------------|---------|
| `-c, --code` | Code file path for boss mode | none |
| `-t, --theme` | Initial theme (claude/codex/opencode) | claude |
| `--speed` | Streaming speed (ms/char) | 25 |
| `-h, --help` | Help | — |

## Dependencies

| Library | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | Style rendering |
| `github.com/charmbracelet/bubbles` | Pre-built components (spinner, etc.) |
| `github.com/spf13/cobra` | CLI argument parsing |
| `github.com/alecthomas/chroma` | Code syntax highlighting |
| `github.com/bmaupin/go-epub` | EPUB parsing |

## Data Flow

```
User starts → Load novel + code file → Welcome page
  → Enter → Reader parses → Pager paginates → Reading mode
      → Space/B: flip pages
      → S: switch Theme (Pager re-paginates with new theme border height)
      → Tab → BossMode → Streamer outputs code char-by-char
          → Tab → back to Reading (preserves current page)
  → Q: quit
```
