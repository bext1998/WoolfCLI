# WoolfCLI 專案狀態

更新日期：2026-05-17

本文件整理目前 repository 的 Phase 1 狀態。依據來源包含 `docs/woolf-spec.md`、`docs/testing.md`、`README.md`、目前 Go package 結構、既有測試與工作樹狀態。本次任務只更新文件，未修改程式碼。

## 專案目前目標

WoolfCLI 的目前目標是完成 Phase 1：一個 Go 實作的多模型 AI 審議 CLI/TUI，透過 OpenRouter 調度多位 Agent，讓文字創作者可以載入草稿、執行多輪審議、保存 session、續接、瀏覽並匯出結果。

最高規格來源仍是 `docs/woolf-spec.md`，特別是：

- 第 5 章 Must Have：多 Agent 流水線、stance tag、role/preset、TUI、SSE、檔案 ingestion、session、OpenRouter、config、成本顯示與 Markdown export。
- 第 7 章：CLI、OpenRouter、Agent、Orchestrator、Ingestion、Session、Export、Cost、Config、TUI 技術規格。
- 第 9 章：`CFG-*`、`API-*`、`ING-*`、`SES-*`、`TUI-*` 等錯誤處理策略。
- 第 11 章與第 14.1 節：測試策略與 Phase 1 驗收矩陣。

## 目前完成狀態

### 已有可驗證基礎

- CLI root command 已註冊 `init`、`start`、`resume`、`list`、`show`、`export`、`fork`、`delete`、`agents`、`config`、`models`。
- `start` command 已能建立 session、讀取 draft、解析 preset/agents，並透過 fake client 測試 orchestrator path，不會在測試中呼叫真實 OpenRouter。
- Agents 模組已內建 6 個 role：`strict-editor`、`casual-reader`、`structure-analyst`、`marketing-eye`、`advocate`、`challenger`。
- Agents 模組已內建 3 組 preset：`editorial`、`brainstorm`、`review`。
- `woolf agents` 已包含 role list/show/add/delete 與 preset list/show 子命令。
- Session store 已有建立、保存、讀取、列表、前綴/索引解析、續接、分叉、刪除與 corrupt JSON 測試。
- Orchestrator pipeline 已有多 agent / 多 round 執行、前序內容組入 context、session persistence、取消後 paused、stream error handling 的測試基礎。
- OpenRouter client 已有 SSE parsing、401/402/429/404/5xx error mapping、retry-after 與 retry 基礎邏輯。
- Markdown exporter 已能輸出 session metadata、source、agents、rounds、stance、tokens 與 summaries。
- CI 已有 GitHub Actions，在 Ubuntu 與 Windows 執行 `go test ./...` 與 `go vet ./...`。
- `scripts/test.ps1` 已作為 Windows PowerShell 預設測試入口，並處理 project-local `GOCACHE` / `GOTMPDIR`。

### 尚未達完整驗收

- TUI 目前仍未形成可驗收的完整互動流程；`internal/tui/app.go` 仍是空 `App` 型別，views/components 需進一步確認與串接。
- Cost tracker 目前只有基本欄位型別，尚未完成 OpenRouter pricing 整合、per-agent/per-round/per-session summary 與 budget warning。
- PDF exporter 目前只有空型別，尚未形成可用輸出流程。
- PDF ingestion 相關檔案存在，但是否符合規格中「具文字層 PDF → Markdown」的可靠驗收仍未確認，需要 regression fixture。
- Ingestion 仍需補 `.md` / `.txt` 錯誤路徑測試，以及掃描 PDF、加密 PDF、無文字層 PDF 的明確 graceful error 測試。
- Orchestrator 還需要補齊 intervention loop、`/skip`、fallback model、多 agent 失敗策略與更完整 state machine。
- Markdown export 雖已有實作，但仍需補對應測試，確認格式與規格第 7.8 節一致。
- README 的狀態描述仍偏「實作前階段」，與目前 repo 已有基礎實作不完全一致。

## 目前主要檔案與責任邊界

- `internal/cli`：目前承擔命令解析與使用者可見流程，不應直接組 OpenRouter request。
- `internal/orchestrator`：目前承擔 Agent 執行順序、context 組裝與 session 保存。
- `internal/openrouter`：目前承擔 API request、SSE parsing、retry 與 API error mapping。
- `internal/session`：目前承擔 session schema 與 file store。
- `internal/ingestion`：目前承擔 `.md`、`.txt`、`.pdf` 讀取與格式轉換入口。
- `internal/exporter`：目前承擔匯出格式，不應改寫 session 決策或內容。
- `internal/tui`：目前應只處理畫面、鍵盤操作與 UI 狀態，不應直接寫 session 檔案。

## 測試狀態

本次文件更新前未修改程式碼。可用的預設驗證命令是：

```powershell
.\scripts\test.ps1
```

目前文件與測試說明顯示此腳本會在 Windows PowerShell 下設定 UTF-8，並在未手動設定時使用 workspace 內的 `.gocache/` 與 `.gotmp/`。

本次任務的必要驗證重點是：

- 三份 Markdown 文件可用 UTF-8 正常讀取。
- Diff 僅限 `docs/PROJECT_BRIEF.md`、`docs/STATUS.md`、`docs/NEXT_ACTIONS.md`。
- 文件不得宣稱未經測試或無法由程式碼確認的功能已完成。

## 目前工作樹注意事項

目前工作樹中已存在與本次任務無關的程式碼與測試腳本變更，例如 `.gitignore`、`scripts/test.ps1`、`internal/agents/registry.go`、`internal/cli/agents.go`、`internal/openrouter/client.go`、`docs/testing.md`。本次文件任務不應回退或修改這些既有變更。

## 可能風險或後續注意事項

- Error code 是使用者可見契約；目前程式採 `API-001` 到 `API-005`，是否正式固定為規格需人工決定。
- Session schema、CLI command output 與 OpenRouter error mapping 都屬外部可觀察行為，後續修改必須同步測試與文件。
- TUI、PDF ingestion、export 與 cost 是目前 Phase 1 主要缺口，若先做局部功能，需避免跨層塞責任。
- 若要更新 README，應明確說明專案已進入基礎實作階段，而不是仍停留在實作前。
