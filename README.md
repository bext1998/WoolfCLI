# Woolf

Woolf 是面向文字創作者與內容工作者的多模型 AI 審議 CLI/TUI。專案目標是透過 OpenRouter API 調度多個 AI Agent，讓不同角色依序閱讀原稿、回應前序觀點，並把討論保存成可續接、可查詢、可匯出的 session。

目前仍在 Phase 1 開發中。CLI、session、agents、OpenRouter 串流、orchestrator 與基礎 ingestion 已有可測骨架；TUI、PDF Phase 1 品質、成本估算與完整匯出流程仍在完善中。

## 目前可用能力

- Cobra CLI 根命令與主要子命令骨架：`init`、`start`、`resume`、`list`、`show`、`export`、`fork`、`delete`、`agents`、`config`、`models`。
- `woolf start` 可讀取 draft、解析 preset / agents，建立 session，並透過 fake 或 OpenRouter client 執行 orchestrator pipeline。
- 內建 6 個 Agent role：`strict-editor`、`casual-reader`、`structure-analyst`、`marketing-eye`、`advocate`、`challenger`。
- 內建 preset：`editorial`、`brainstorm`、`critique`、`review`。
- `woolf agents list | show | add | delete | validate` 可管理內建與自訂 YAML role。
- OpenRouter client 支援 Chat Completions SSE 串流、usage 解析、API 錯誤碼映射，以及 429 / 5xx retry。
- Orchestrator 會依序執行多 Agent，保存每個回覆、usage token、session totals，並在後續 Agent context 中包含前序完整討論。
- Context builder 會組裝 system prompt、focus areas、response template、draft、session summaries、user interventions、focus range 與 previous discussion。
- Session store 支援建立、儲存、讀取、續接、fork、刪除與列表。
- Ingestion 已有 `.md`、`.txt` 與 PDF Phase 1 入口。

## 快速使用

先設定 OpenRouter API key：

```powershell
$env:OPENROUTER_API_KEY = "sk-or-..."
```

建立並執行一輪 editorial preset：

```powershell
go run .\cmd\woolf --config .\tmp\config.toml start --draft .\chapter3.md --preset editorial --rounds 1
```

列出 Agent：

```powershell
go run .\cmd\woolf --config .\tmp\config.toml agents list
```

查看單一 Agent：

```powershell
go run .\cmd\woolf --config .\tmp\config.toml agents show strict-editor
```

加入自訂 Agent YAML：

```powershell
go run .\cmd\woolf --config .\tmp\config.toml agents add .\my-agent.yaml
```

## 專案結構

- `cmd/woolf`：CLI 程式入口。
- `internal/cli`：Cobra 指令、參數解析與使用者可見錯誤。
- `internal/tui`：Bubble Tea TUI 骨架、view 與元件。
- `internal/orchestrator`：Agent pipeline、context builder、intervention 與流程狀態。
- `internal/agents`：Agent role、preset、YAML 載入與 registry。
- `internal/openrouter`：OpenRouter API client、SSE 串流、模型列表、rate limit 與錯誤映射。
- `internal/ingestion`：`.md`、`.txt`、`.pdf` 讀取與轉換入口。
- `internal/session`：Session schema、store、resume、fork、search 與持久化。
- `internal/exporter`：Markdown / PDF 匯出骨架。
- `internal/cost`：token 與費用估算骨架。
- `internal/config`：TOML 設定、預設值、路徑與環境變數載入。
- `pkg/pdfparse`：未來公開復用的 PDF parsing 型別或工具。

## 開發與測試

本專案使用 Go 1.22 或更新版本。Windows PowerShell 預設測試命令：

```powershell
.\scripts\test.ps1
```

可選檢查：

```powershell
.\scripts\test.ps1 -Vet
.\scripts\test.ps1 -Race
.\scripts\test.ps1 -Coverage
```

更多測試說明見 [docs/testing.md](docs/testing.md)，Phase 1 進度見 [docs/todo.md](docs/todo.md)，產品規格以 [docs/woolf-spec.md](docs/woolf-spec.md) 為準。

## 安全原則

- 不硬編 API key、token、密碼或任何敏感資訊。
- OpenRouter API key 只從環境變數或本機設定讀取。
- CLI、TUI、log 與 config show 顯示 key 時必須遮蔽。
- 測試不得呼叫真實 OpenRouter，需使用 fake client、mock transport 或 fixture。
- Session 可能包含使用者原稿，應只保存在本機 runtime data 目錄。

## License

Woolf 使用 Apache License 2.0。詳見 [LICENSE](LICENSE)。
