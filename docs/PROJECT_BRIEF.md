# WoolfCLI Project Brief

更新日期：2026-05-19

本文件根據目前 repository、`docs/spec.md`、`docs/testing.md`、`README.md` 與現有 Go 程式碼整理。無法由目前檔案直接確認的內容一律標記為「未確認」。

## 專案目前目標

WoolfCLI 是面向文字創作者與內容工作者的多模型 AI 審議 CLI/TUI。它不是通用聊天機器人、IDE agent、RAG 平台或 Web app；核心目標是讓使用者投入草稿後，由多位 AI Agent 依序閱讀稿件與前序討論，提出不同立場的回饋，並形成可保存、可續接、可匯出的本機 session。

Phase 1 目標以 `docs/spec.md` 第 5 章 Must Have 與第 14.1 節驗收矩陣為準：

- 以 Go 實作 CLI/TUI，CLI 使用 Cobra，TUI 使用 Bubble Tea 系列套件。
- 透過 OpenRouter Chat Completions / SSE 調度多模型 Agent。
- 支援 2 到 6 位 Agent 依序發言，後發言者可讀取前序完整內容。
- 支援 `agree`、`disagree`、`extend`、`neutral` 的 response stance tag。
- 內建至少 6 個 Agent role，並支援 YAML 自訂 role。
- 提供至少 3 組 preset，一鍵啟動不同審議情境。
- 支援 `.md`、`.txt`、具文字層 `.pdf` ingestion。
- 以本機 JSON session 作為持久化單位，支援建立、讀取、續接、列表、分叉與匯出。
- 支援 Markdown 匯出；PDF 匯出在規格中屬 Should Have / Phase 2 方向，當前是否列入近期實作仍未確認。
- TUI 應服務長文創作流程，包含討論串、輸入區、狀態列、鍵盤操作、串流顯示、成本提示與過小視窗 fallback。

## 目前架構

目前 repository 已建立主要 Go package 與模組邊界：

- `cmd/woolf`：CLI binary 入口。
- `internal/cli`：Cobra 指令、參數解析、命令流程與使用者可見輸出。
- `internal/agents`：內建 role、preset、YAML 載入與 registry。
- `internal/orchestrator`：Agent pipeline、事件、context 組裝、取消與 session 保存。
- `internal/openrouter`：OpenRouter client、SSE stream parsing、API error mapping、retry / rate limit。
- `internal/ingestion`：Markdown、text、PDF ingestion 入口。
- `internal/ingestion/pdf`：PDF 文字層解析子模組。
- `internal/session`：session schema、file store、resume、fork、search。
- `internal/exporter`：Markdown exporter 與 PDF exporter 型別。
- `internal/cost`：pricing / tracker 型別。
- `internal/config`：TOML config、runtime path、環境變數與 API key masking。
- `internal/tui`：Bubble Tea 相關 app、view、component 檔案。
- `pkg/pdfparse`：公開 PDF parsing 型別或工具入口。

專案文件目前以 `docs/` 作為狀態文件集中位置；`docs/spec.md` 是最高產品規格來源，`docs/PROJECT_BRIEF.md`、`docs/STATUS.md` 與 `docs/NEXT_ACTIONS.md` 是後續任務交接文件。

## 目前完成狀態摘要

- CLI 指令面已包含 `init`、`start`、`resume`、`list`、`show`、`export`、`fork`、`delete`、`agents`、`config`、`models`。
- `start` 已可建立 session、載入 draft、解析 preset/agents，並透過 orchestrator pipeline 執行 agent 串流流程。
- Agents 模組已內建 6 個 role 與 3 組 preset，並支援自訂 YAML role 載入與 registry 查詢。
- Session 模組已有 JSON schema、建立、保存、讀取、列表、續接、分叉、刪除與 corrupt JSON 測試。
- OpenRouter client 已有串流解析、HTTP error mapping、429 retry-after 與 5xx retry 基礎邏輯。
- Markdown exporter 已能輸出 session metadata、source、agents、rounds、stance、tokens 與 summaries。
- 規格引用已統一到 `docs/spec.md`，並移除根目錄舊版 `STATUS.md` / `NEXT_ACTIONS.md`，避免狀態文件分叉。
- TUI、cost tracker、PDF exporter 目前更接近骨架或初步實作，尚未達到 Phase 1 完整驗收。
- PDF ingestion 是否已符合「具文字層 PDF 基本解析」仍需以 regression fixture 驗證；目前文件應視為未確認。

## 產品邊界

目前不應納入的方向：

- Web UI、雲端同步、多人協作、資料庫後端或外部帳號系統。
- 非 OpenRouter provider。
- 長期跨 session agent memory。
- OCR、掃描 PDF、加密 PDF 或複雜版面還原。
- Phase 2 功能如 session diff、重播模式、分歧度分析、沙龍記憶，除非使用者明確要求。

## 未確認事項

- PDF ingestion 對真實文字層 PDF 的覆蓋率與 regression fixture 狀態未確認。
- PDF exporter 是否要進入近期範圍未確認。
- TUI 應先完成完整互動模式，或先讓 CLI-only pipeline 達到可驗收狀態，需人工決定。
- 是否正式以 `API-001` 到 `API-005` 作為對外 OpenRouter/API error code 契約，需人工決定。
