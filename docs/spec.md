# Woolf — 正式規格書 v1.0

## 多模型 AI 審議 CLI：為文字工作者設計的創作夥伴

> **命名由來**：取自 Virginia Woolf 與她所參與的布魯姆斯伯里文學沙龍——一群作家、藝術家定期聚會、互相批評與激辯的傳統。Woolf 的核心精神，是把這種多元視角的文學審議帶進終端機。

---

**文件版本**：1.0
**最後更新**：2026-05-08
**狀態**：Draft → Review
**授權**：待定（考慮開源）

---

## 目錄

1. [術語表](#1-術語表)
2. [專案概述](#2-專案概述)
3. [問題與目標](#3-問題與目標)
4. [使用者畫像與情境](#4-使用者畫像與情境)
5. [功能規格](#5-功能規格)
6. [系統架構](#6-系統架構)
7. [技術規格](#7-技術規格)
   - 7.1 [CLI 指令規格](#71-cli-指令規格)
   - 7.2 [OpenRouter Client](#72-openrouter-client)
   - 7.3 [Agent 角色系統](#73-agent-角色系統)
   - 7.4 [流水線 Orchestrator（狀態機）](#74-流水線-orchestrator狀態機)
   - 7.5 [Context 管理](#75-context-管理)
   - 7.6 [Ingestion 模組](#76-ingestion-模組)
   - 7.7 [Session 管理](#77-session-管理)
   - 7.8 [匯出模組](#78-匯出模組)
   - 7.9 [成本追蹤](#79-成本追蹤)
   - 7.10 [Configuration](#710-configuration)
   - 7.11 [TUI 規格](#711-tui-規格)
8. [資料格式定義](#8-資料格式定義)
9. [錯誤處理策略](#9-錯誤處理策略)
10. [安全性考量](#10-安全性考量)
11. [測試策略](#11-測試策略)
12. [安裝與發佈](#12-安裝與發佈)
13. [風險與緩解](#13-風險與緩解)
14. [KPI 與驗收標準](#14-kpi-與驗收標準)
15. [Phase 2 路線圖](#15-phase-2-路線圖)
16. [創意延伸](#16-創意延伸)
17. [附錄](#17-附錄)

---

## 1. 術語表

| 術語 | 定義 |
|------|------|
| **Agent** | 一個綁定特定模型與角色定義的 AI 參與者。每個 Agent 由 model + role + system prompt 組成。 |
| **Session** | 一次完整的審議過程，從載入稿件到結束討論。Session 為持久化單位。 |
| **Round（輪次）** | 所有 Agent 各發言一次為一輪。一個 Session 可包含多輪。 |
| **Pipeline（流水線）** | Agent 依序發言的排程機制。後發言者可讀取前發言者的完整內容。 |
| **Stance（立場標籤）** | Agent 對前序發言者表態的標記：`agree`、`disagree`、`extend`、`neutral`。 |
| **Intervention（使用者介入）** | 使用者在 Round 之間插入的指令、追問或補充材料。 |
| **Preset（預設組合）** | 預先定義好的 Agent 組合，適用於特定審議場景。 |
| **Ingestion** | 將外部檔案（.md / .pdf / .txt）載入系統的前處理流程。 |
| **Context Window** | 模型單次呼叫可接收的最大 token 數量。 |
| **Micro-session** | 針對稿件特定段落的小型審議。 |
| **Summarizer** | 流水線末端的摘要 Agent，負責產出結構化審議總結。 |
| **OpenRouter** | 第三方 API 聚合服務，提供統一介面呼叫多家 LLM。 |
| **TUI** | Terminal User Interface，終端機圖形介面。 |

---

## 2. 專案概述

Woolf 是一套面向文字創作者與工作者的 TUI 應用程式，透過 OpenRouter API 同時調度多個 AI 模型，模擬「沙龍式審議流水線」。使用者可投入既有稿件或從零起草，由多位 AI 角色依序發言、互相回應、互相辯駁，協助創作者在最短時間內取得多元視角的回饋。

**核心價值主張**：單一 AI 只能給你一種聲音。Woolf 讓你同時聽到嚴格的編輯、挑剔的讀者、支持你的辯護者——然後自己做決定。

**技術定位**：Go 語言 CLI/TUI 應用 → OpenRouter API → 多模型串流 → 本地 Session 持久化。

**目標平台**：macOS、Linux、Windows（WSL / Windows Terminal）。

---

## 3. 問題與目標

### 3.1 待解決問題

| # | 問題 | 影響 |
|---|------|------|
| P1 | 單一 AI 回饋視角單一，難以揭露創作盲點 | 創作品質瓶頸 |
| P2 | 缺乏低摩擦的多元回饋工具 | 創作者依賴人工審稿，週期長 |
| P3 | Web AI 介面不適合長時間沉浸式創作工作流 | 鍵盤流中斷、不可程式化 |
| P4 | 多模型工具普遍缺乏角色定位與互相回應機制 | 回饋缺乏結構性 |
| P5 | 審議歷史無法持久化與回溯 | 無法追蹤修改脈絡 |

### 3.2 目標層級

| 層級 | 目標 | 衡量方式 |
|------|------|----------|
| **核心** | 穩定可運作的多模型流水線審議系統 | 3 Agent × 3 輪完整跑完不出錯 |
| **體驗** | TUI 操作流暢，鍵盤友善，長文不卡頓 | 5000 字輸入無凍結 |
| **資料** | Session 自動存檔、可索引、可重播 | 中斷後可續接 |
| **彈性** | 模型、角色、流程皆可使用者自訂 | 自訂 YAML 角色可載入 |
| **透明** | Token 用量與費用即時可見 | 與 OpenRouter Dashboard 誤差 < 5% |
| **可維護** | 模組清晰、可測試、可擴展 | 單元測試覆蓋率 ≥ 70% |

---

## 4. 使用者畫像與情境

### 4.1 Personas

**Persona A — 長篇小說作者（小婷）**
- 背景：正在寫長篇小說，已有數章草稿
- 痛點：缺乏可信的多視角審稿者；朋友審稿太慢、太客氣
- 使用情境：每完成一章就丟進 Woolf，讓「編輯」和「讀者」同時給意見
- 成功指標：在 10 分鐘內取得可行動的修改建議

**Persona B — 專欄作者（阿凱）**
- 背景：每週產出 2-3 篇長文，需要快速回饋循環
- 痛點：Web AI 每次只能問一個模型，切換麻煩
- 使用情境：寫完草稿後，用 Woolf 同時取得邏輯、結構、行銷三個角度的意見
- 成功指標：回饋流程從 30 分鐘縮短到 5 分鐘

**Persona C — 內容編輯（Lena）**
- 背景：評估他人投稿，需要客觀多元意見作為決策參考
- 痛點：主觀判斷需要佐證，但沒有第二個編輯可以討論
- 使用情境：將投稿丟進 Woolf，取得多角度評估後再給作者回饋
- 成功指標：決策信心提升，回饋品質更客觀

### 4.2 核心 User Flow

```
[啟動]
  $ woolf start --draft chapter3.md --preset editorial
         │
         ▼
[Session 初始化]
  ├─ 驗證 API key
  ├─ 載入稿件
  │   ├─ .md → 直讀
  │   ├─ .pdf → ingestion/pdf 轉換
  │   └─ .txt → 直讀
  ├─ 載入 Agent Preset 或互動式選擇
  ├─ 顯示 Session 資訊面板（稿件預覽、Agent 清單、預估成本）
  └─ 使用者確認 → 進入 TUI
         │
         ▼
[TUI 主畫面 — 等待指令]
  ├─ 使用者輸入開場問題或指令（選填）
  └─ 按 Enter 或輸入 `/start` → 啟動流水線
         │
         ▼
[審議流水線 — Round 1]
  ┌─ Orchestrator 依序調度：
  │   Agent A → 讀取 [系統 prompt + 原稿 + 使用者開場] → 串流回應
  │   Agent B → 讀取 [系統 prompt + 原稿 + 使用者開場 + A 回應] → 串流回應
  │   Agent C → 讀取 [系統 prompt + 原稿 + 使用者開場 + A + B 回應] → 串流回應
  └─ Round 1 完成 → 自動儲存
         │
         ▼
[使用者介入點]
  ├─ `/next`           → 繼續下一輪
  ├─ 直接輸入文字       → 追問或補充，作為下一輪 context
  ├─ `/focus 12-18`    → 下一輪聚焦稿件第 12-18 行
  ├─ `/add-file x.md`  → 補充材料
  ├─ `/skip <agent>`   → 下一輪跳過某 Agent
  ├─ `/summarize`      → 立即觸發摘要 Agent
  └─ `/end`            → 結束審議
         │
         ▼
[Round 2 ... N]
  ├─ 同上邏輯，context 累積
  ├─ 若達到 max_rounds → 自動進入結束流程
  └─ 每輪結束自動儲存
         │
         ▼
[結束流程]
  ├─ Summarizer Agent（若啟用）產出結構化總結
  ├─ 顯示 Session 統計（總 token、總費用、輪數）
  ├─ 詢問是否匯出（.md / .pdf）
  └─ Session 標記為 completed
```

### 4.3 次要 User Flow

**Flow B — 從零創作**
```
$ woolf start --preset brainstorm
> 使用者輸入：「我想寫一個關於時間旅行的短篇，主角是一個鐘錶匠」
→ Agent A（創意發想）展開構思
→ Agent B（結構分析）提出架構建議
→ Agent C（讀者視角）指出潛在吸引力與風險
→ 使用者根據回饋細化方向 → 繼續迭代
```

**Flow C — 歷史回顧**
```
$ woolf list
  [1] 20260507-143022 第三章草稿審議 (completed, 3 rounds)
  [2] 20260506-091500 開頭段落打磨  (completed, 2 rounds)
  [3] 20260505-200000 角色設定討論  (paused, 1 round)

$ woolf show 20260507-143022
  → TUI 顯示完整討論記錄（唯讀模式）

$ woolf resume 20260505-200000
  → 從暫停處繼續
```

**Flow D — 續接分叉**
```
$ woolf fork 20260507-143022 --title "第三章-改寫版審議"
  → 複製原 session，新 session 載入修改後稿件
  → 使用相同 Agent 組合重新審議
  → 之後可用 woolf diff 比較兩次審議
```

---

## 5. 功能規格

### 🔴 Must Have（Phase 1 必須完成）

| ID | 功能 | 描述 |
|----|------|------|
| **M-01** | 多模型流水線 | 2–6 Agent 依序發言，後者讀取前者完整內容 |
| **M-02** | 立場標籤機制 | 每位 Agent 可對前序發言標記 agree / disagree / extend / neutral |
| **M-03** | Agent 角色系統 | 內建 ≥ 6 角色模板，支援 YAML 自訂 |
| **M-04** | Agent 預設組合 | 至少 3 組 Preset，一鍵啟動 |
| **M-05** | TUI 三區佈局 | 討論串 / 輸入區 / 狀態列，Bubble Tea 實作 |
| **M-06** | 串流顯示 | SSE 串流回應逐字渲染至 TUI |
| **M-07** | 鍵盤全操作 | 所有功能可純鍵盤操作 |
| **M-08** | .md 讀取 | 直接讀入 .md 檔案 |
| **M-09** | .txt 讀取 | 直接讀入 .txt 檔案 |
| **M-10** | .pdf 讀取（自建） | 純文字層 PDF → Markdown 轉換（自建解析器） |
| **M-11** | Session 自動儲存 | JSON 格式，每輪結束後自動寫入 |
| **M-12** | Session 續接 | 中斷後可從上次位置繼續 |
| **M-13** | Session 瀏覽 | CLI 列出歷史 sessions |
| **M-14** | OpenRouter 整合 | API key 設定、模型清單查詢、串流對話 |
| **M-15** | 設定檔 | TOML 格式，管理 API key、預設值、Agent 定義 |
| **M-16** | 使用者介入 | 在輪次之間插入指令、追問、補充 |
| **M-17** | 成本即時顯示 | TUI 狀態列顯示 token 與費用 |
| **M-18** | .md 匯出 | 將完整 session 匯出為格式化 Markdown |

### 🟠 Should Have（Phase 1 建議具備）

| ID | 功能 | 描述 |
|----|------|------|
| **S-01** | Session TUI 瀏覽器 | TUI 內互動式瀏覽歷史 sessions |
| **S-02** | Context 摘要壓縮 | 超過 token 閾值時自動摘要舊輪次 |
| **S-03** | 預算上限警告 | 達到設定預算閾值時 TUI 提示 |
| **S-04** | 焦點模式 | `/focus` 指定段落，下一輪僅針對該段落 |
| **S-05** | Session 分叉 | 複製既有 session 為新 session，載入修改版稿件 |
| **S-06** | .pdf 匯出 | 將 session 匯出為排版過的 PDF |
| **S-07** | 輸入語法高亮 | 輸入區支援基本 Markdown 語法高亮 |

### 🟢 Could Have（Phase 2 以後）

| ID | 功能 | 描述 |
|----|------|------|
| **C-01** | Summarizer Agent | 流水線末端自動產出結構化總結 |
| **C-02** | 重播模式 | 逐字串流重播歷史 session |
| **C-03** | Session diff | 比較同稿件兩次審議結果 |
| **C-04** | 沙龍記憶 | Agent 跨 Session 累積對作者的觀察 |
| **C-05** | Micro-session | 針對特定段落啟動小型審議 |
| **C-06** | 分歧度指標 | 視覺化 Agent 共識 / 爭議程度 |
| **C-07** | 辯論模式 | 兩位 Agent 正反辯論 + 裁判 |
| **C-08** | 本地模型支援 | Ollama 整合 |
| **C-09** | 多語系介面 | en / zh-TW 切換 |
| **C-10** | 雲端同步 | Session 遠端備份（選用） |
| **C-11** | 插件系統 | 第三方 Agent 角色定義分享 |

---

## 6. 系統架構

### 6.1 模組總覽

```
woolf/
├── cmd/woolf/                  # CLI 入口
│   └── main.go                 # cobra root command
│
├── internal/
│   ├── cli/                    # Subcommand 定義
│   │   ├── start.go
│   │   ├── resume.go
│   │   ├── list.go
│   │   ├── show.go
│   │   ├── delete.go
│   │   ├── fork.go
│   │   ├── export.go
│   │   ├── agents.go
│   │   ├── models.go
│   │   ├── config.go
│   │   └── init.go
│   │
│   ├── tui/                    # TUI 渲染層
│   │   ├── app.go              # 主 Bubble Tea Model
│   │   ├── keymap.go           # 快捷鍵定義
│   │   ├── theme.go            # 色彩主題
│   │   ├── views/
│   │   │   ├── discussion.go   # 討論串面板
│   │   │   ├── input.go        # 輸入面板
│   │   │   ├── status.go       # 狀態列
│   │   │   ├── preview.go      # 稿件預覽
│   │   │   └── session_browser.go
│   │   └── components/
│   │       ├── agent_badge.go  # Agent 標識元件
│   │       ├── stance_tag.go   # 立場標籤
│   │       ├── cost_meter.go   # 費用計
│   │       └── progress.go     # 進度指示
│   │
│   ├── orchestrator/           # 流水線排程核心
│   │   ├── pipeline.go         # Pipeline 狀態機
│   │   ├── state.go            # 狀態定義
│   │   ├── context_builder.go  # Context 組裝
│   │   ├── compressor.go       # Context 壓縮
│   │   └── intervention.go     # 使用者介入處理
│   │
│   ├── agents/                 # Agent 角色系統
│   │   ├── role.go             # 角色結構
│   │   ├── prompt.go           # Prompt 模板引擎
│   │   ├── loader.go           # YAML / TOML 載入
│   │   ├── preset.go           # 預設組合
│   │   └── builtin/            # 內建角色定義
│   │       ├── strict-editor.yaml
│   │       ├── casual-reader.yaml
│   │       ├── structure-analyst.yaml
│   │       ├── marketing-eye.yaml
│   │       ├── advocate.yaml
│   │       └── challenger.yaml
│   │
│   ├── openrouter/             # OpenRouter API Client
│   │   ├── client.go           # HTTP client
│   │   ├── stream.go           # SSE 串流處理
│   │   ├── models.go           # 模型清單與定價
│   │   ├── errors.go           # 錯誤類型
│   │   └── ratelimit.go        # Rate limit 處理
│   │
│   ├── ingestion/              # 檔案讀取
│   │   ├── ingester.go         # 統一介面
│   │   ├── md.go
│   │   ├── txt.go
│   │   └── pdf/                # 自建 PDF 解析
│   │       ├── lexer.go        # Token 解析
│   │       ├── parser.go       # 物件樹建構
│   │       ├── xref.go         # Cross-reference table
│   │       ├── stream.go       # Stream 解碼（FlateDecode 等）
│   │       ├── extractor.go    # 文字流提取
│   │       ├── encoder.go      # 字元編碼處理
│   │       ├── cmap.go         # CMap / ToUnicode 解析
│   │       └── markdown.go     # Text → Markdown 轉換
│   │
│   ├── session/                # Session 管理
│   │   ├── store.go            # 檔案系統讀寫
│   │   ├── schema.go           # JSON 結構定義
│   │   ├── resume.go           # 續接邏輯
│   │   ├── fork.go             # 分叉邏輯
│   │   └── search.go           # 搜尋 / 篩選
│   │
│   ├── exporter/               # 匯出
│   │   ├── exporter.go         # 統一介面
│   │   ├── markdown.go         # .md 匯出
│   │   └── pdf.go              # .pdf 匯出（Phase 2）
│   │
│   ├── cost/                   # 費用追蹤
│   │   ├── tracker.go          # 計算邏輯
│   │   └── pricing.go          # OpenRouter 定價快取
│   │
│   └── config/                 # 設定
│       ├── config.go           # TOML 解析
│       ├── defaults.go         # 預設值
│       └── paths.go            # 路徑常數
│
├── pkg/                        # 對外可重用 package（未來）
│   └── pdfparse/               # PDF 解析器獨立封裝
│
├── testdata/                   # 測試用檔案
│   ├── sample.md
│   ├── sample.pdf
│   └── sessions/
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

### 6.2 依賴關係圖

```
                 ┌──────────┐
                 │ cmd/woolf│
                 └────┬─────┘
                      │
            ┌─────────┴──────────┐
            ▼                    ▼
       ┌────────┐          ┌─────────┐
       │  cli   │          │   tui   │
       └───┬────┘          └────┬────┘
           │                    │
           └────────┬───────────┘
                    ▼
            ┌──────────────┐
            │ orchestrator │ ◄─── 流水線核心
            └──────┬───────┘
                   │
       ┌───────────┼───────────┐
       ▼           ▼           ▼
  ┌─────────┐ ┌────────┐ ┌─────────┐
  │ agents  │ │ openr. │ │ session │
  └─────────┘ └────────┘ └─────────┘
                               │
       ┌───────────────────────┤
       ▼                       ▼
  ┌───────────┐          ┌──────────┐
  │ ingestion │          │ exporter │
  └───────────┘          └──────────┘
       │
       ▼
  ┌─────────┐      ┌──────┐     ┌────────┐
  │ pdf/    │      │ cost │     │ config │
  └─────────┘      └──────┘     └────────┘

  ※ 箭頭方向 = 依賴方向（上層依賴下層）
  ※ 不可出現循環依賴
```

### 6.3 模組職責矩陣

| 模組 | 職責 | 依賴 | 被依賴 |
|------|------|------|--------|
| `cmd/woolf` | 程式入口 | cli, tui | — |
| `cli` | Subcommand 解析、參數驗證 | orchestrator, session, agents, config | cmd |
| `tui` | 畫面渲染、事件處理 | orchestrator, session, cost, config | cmd |
| `orchestrator` | 流水線排程、context 組裝 | agents, openrouter, session | cli, tui |
| `agents` | 角色定義、prompt 組裝 | config | orchestrator |
| `openrouter` | API 通訊、串流 | config | orchestrator, cost |
| `ingestion` | 檔案前處理 | pdf/ | session, cli |
| `pdf/` | PDF 解析 | — | ingestion |
| `session` | 持久化、續接、分叉 | ingestion | orchestrator, cli, tui, exporter |
| `exporter` | 匯出 .md / .pdf | session | cli |
| `cost` | 費用計算 | openrouter | tui |
| `config` | 設定檔解析 | — | 幾乎所有模組 |

---

## 7. 技術規格

### 7.1 CLI 指令規格

**指令框架**：[cobra](https://github.com/spf13/cobra)

```
woolf <command> [subcommand] [flags]

Commands:
  init                                # 初始化設定（互動式）
  start     [--draft FILE] [--preset NAME] [--agents a,b,c] [--rounds N]
  resume    <session-id>
  list      [--limit N] [--since DATE] [--status STATUS]
  show      <session-id>              # 唯讀檢視
  delete    <session-id> [--force]
  fork      <session-id> [--draft FILE] [--title TITLE]
  export    <session-id> --format md|pdf [--output PATH]
  agents    list | show <name> | add | edit <name> | delete <name>
  agents    preset list | preset show <name>
  models    [--pricing]
  config    show | edit | reset
  version

Global Flags:
  --config PATH     # 指定設定檔路徑（覆蓋預設）
  --verbose         # 詳細輸出
  --no-color        # 停用顏色
  --debug           # 除錯模式
```

**指令詳細規格**

| 指令 | 必要參數 | 選填參數 | 行為 |
|------|----------|----------|------|
| `start` | — | `--draft`, `--preset`, `--agents`, `--rounds` | 建立新 session，進入 TUI |
| `resume` | `session-id` | — | 載入暫停的 session，進入 TUI |
| `list` | — | `--limit`(預設 20), `--since`, `--status` | 列出 sessions（表格格式） |
| `show` | `session-id` | — | 在 TUI 唯讀模式顯示完整記錄 |
| `delete` | `session-id` | `--force` | 刪除 session（無 --force 需確認） |
| `fork` | `session-id` | `--draft`, `--title` | 複製 session 為新 session |
| `export` | `session-id` | `--format`(必要), `--output` | 匯出檔案 |

**session-id 解析規則**：
- 完整 ID：`20260507-143022-chapter3`
- 前綴匹配：`2026050` → 若唯一則自動匹配，若多筆則列出選項
- 數字索引：`1`, `2` → 對應 `woolf list` 的序號

### 7.2 OpenRouter Client

**Go Interface 定義**

```go
// openrouter/client.go

type Client interface {
    // ListModels 查詢可用模型清單
    ListModels(ctx context.Context) ([]Model, error)

    // StreamChat 發起串流對話，回傳 channel
    StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)

    // GetModelPricing 取得特定模型定價
    GetModelPricing(modelID string) (*Pricing, error)
}

type ChatRequest struct {
    Model       string     `json:"model"`
    Messages    []Message  `json:"messages"`
    Stream      bool       `json:"stream"`
    Temperature float64    `json:"temperature,omitempty"`
    MaxTokens   int        `json:"max_tokens,omitempty"`
    TopP        float64    `json:"top_p,omitempty"`
}

type Message struct {
    Role    string `json:"role"`    // "system" | "user" | "assistant"
    Content string `json:"content"`
}

type StreamEvent struct {
    Type    StreamEventType
    Delta   string    // 增量文字
    Usage   *Usage    // 最後一個 event 附帶
    Error   error     // 若非 nil，表示串流錯誤
    Done    bool      // 串流結束
}

type StreamEventType int
const (
    EventDelta StreamEventType = iota
    EventDone
    EventError
)

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}

type Model struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    ContextLen  int     `json:"context_length"`
    Pricing     Pricing `json:"pricing"`
}

type Pricing struct {
    PromptCostPerToken     float64
    CompletionCostPerToken float64
}
```

**API 端點**

| 方法 | URL | 用途 |
|------|-----|------|
| GET | `https://openrouter.ai/api/v1/models` | 模型清單 |
| POST | `https://openrouter.ai/api/v1/chat/completions` | 串流對話 |

**HTTP Headers**

```
Authorization: Bearer <OPENROUTER_API_KEY>
Content-Type: application/json
HTTP-Referer: https://github.com/woolf-cli  (選填，OpenRouter 建議)
X-Title: Woolf                               (選填)
```

**SSE 串流解析**

```
Event stream format:
  data: {"id":"...","choices":[{"delta":{"content":"文字"}}]}
  data: {"id":"...","choices":[{"delta":{},"finish_reason":"stop"}],"usage":{...}}
  data: [DONE]

解析規則：
  1. 逐行讀取，過濾空行
  2. 去掉 "data: " prefix
  3. 若為 "[DONE]" → 串流結束
  4. JSON parse → 提取 delta.content
  5. 最後一個有效 event 通常帶 usage → 提取 token 統計
```

**重試策略**

| 狀況 | 行為 |
|------|------|
| HTTP 429 (Rate Limit) | 依 Retry-After header 等待，最多重試 3 次 |
| HTTP 5xx | 指數退避（1s, 2s, 4s），最多 3 次 |
| 網路逾時 | 依 config.timeout_seconds，逾時後中斷並通知 TUI |
| 模型不可用 | 標記該 Agent 為 `skipped`，流水線繼續 |

### 7.3 Agent 角色系統

**Go Interface 定義**

```go
// agents/role.go

type Role struct {
    Name         string   `yaml:"name"`
    DisplayName  string   `yaml:"display_name"`
    Model        string   `yaml:"model"`
    Stance       Stance   `yaml:"stance"`       // critique | support | neutral
    Temperature  float64  `yaml:"temperature"`
    FocusAreas   []string `yaml:"focus_areas"`
    SystemPrompt string   `yaml:"system_prompt"`
    ResponseTmpl string   `yaml:"response_template"`
    Color        string   `yaml:"color"`         // TUI 顏色識別
}

type Stance string
const (
    StanceCritique Stance = "critique"
    StanceSupport  Stance = "support"
    StanceNeutral  Stance = "neutral"
)

type Preset struct {
    Name        string   `yaml:"name"`
    DisplayName string   `yaml:"display_name"`
    Description string   `yaml:"description"`
    Agents      []string `yaml:"agents"`  // 角色 name 列表
}

// AgentRegistry 管理所有已載入角色
type AgentRegistry interface {
    Get(name string) (*Role, error)
    List() []Role
    ListPresets() []Preset
    LoadCustom(path string) error
    Register(role Role) error
}
```

**內建角色定義（6 個）**

```yaml
# strict-editor.yaml
name: strict-editor
display_name: 嚴格編輯
model: openai/gpt-4o
stance: critique
temperature: 0.5
color: "#E74C3C"
focus_areas:
  - 結構邏輯
  - 用詞精準
  - 段落銜接
  - 語病與冗贅
system_prompt: |
  你是一位資深文學編輯，以嚴格和精準著稱。你的任務是：
  1. 直接點出稿件最關鍵的 1-3 個結構或邏輯問題
  2. 標出不精準或可改善的用詞（引用原文行號或句子）
  3. 若前序討論中有其他 Agent 的觀點，明確表態你同意或反對，並說明理由
  4. 對每個問題提出具體修改方向（不只是「要改」，而是「怎麼改」）

  格式要求：
  - 先表態你對前序意見的立場（若有）
  - 再依重要度列出你的觀察
  - 最後給出優先修改建議
response_template: |
  ## 對前序討論的回應
  {responses_to_others}

  ## 關鍵問題
  {observations}

  ## 優先修改方向
  {suggestions}
```

```yaml
# casual-reader.yaml
name: casual-reader
display_name: 一般讀者
model: anthropic/claude-sonnet-4
stance: neutral
temperature: 0.7
color: "#3498DB"
focus_areas:
  - 閱讀體驗
  - 情感共鳴
  - 節奏感受
  - 理解障礙
system_prompt: |
  你是一位認真的普通讀者，沒有專業文學背景但閱讀量豐富。你的任務是：
  1. 分享你的第一手閱讀感受（哪裡被吸引、哪裡開始走神）
  2. 標出你覺得困惑或斷裂的段落
  3. 對其他 Agent 的技術性批評，從讀者感受出發表達同意或不同意
  4. 如果你覺得某段寫得很好，也要明確說出來

  你的語氣是真誠、直覺、非技術性的。
```

```yaml
# structure-analyst.yaml
name: structure-analyst
display_name: 結構分析師
model: google/gemini-2.5-pro
stance: neutral
temperature: 0.4
color: "#2ECC71"
focus_areas:
  - 敘事弧線
  - 段落佈局
  - 開頭與結尾
  - 節奏與張力
system_prompt: |
  你是一位敘事結構專家。你的分析聚焦在文本的架構層面：
  1. 分析整體敘事弧線（起承轉合或其他結構）
  2. 評估段落排列是否最佳化（有無該合併、拆分、移動的段落）
  3. 檢視開頭是否有效抓住注意力，結尾是否有力
  4. 對其他 Agent 的意見，從結構角度補充或反駁
```

```yaml
# marketing-eye.yaml
name: marketing-eye
display_name: 行銷視角
model: openai/gpt-4o
stance: neutral
temperature: 0.6
color: "#F39C12"
focus_areas:
  - 標題吸引力
  - 目標受眾匹配
  - 傳播潛力
  - 開頭 hook
system_prompt: |
  你從行銷與傳播的角度審視文本：
  1. 評估標題或開頭是否具有點擊吸引力
  2. 分析目標受眾是否明確，語調是否匹配
  3. 指出最具「可分享性」的段落或觀點
  4. 建議如何強化傳播力而不犧牲內容品質
```

```yaml
# advocate.yaml
name: advocate
display_name: 辯護者
model: anthropic/claude-sonnet-4
stance: support
temperature: 0.6
color: "#9B59B6"
focus_areas:
  - 作者意圖理解
  - 優點放大
  - 反駁批評
system_prompt: |
  你的角色是作者的辯護者。你的任務是：
  1. 先理解並闡述你認為作者的創作意圖
  2. 指出稿件中做得好的地方，並解釋為什麼好
  3. 針對其他 Agent 的批評，站在作者立場提出反駁
  4. 若你認為某個批評確實有道理，也承認，但提出兼顧作者意圖的折衷方案

  注意：你不是無條件吹捧。你是有理有據的辯護。
```

```yaml
# challenger.yaml
name: challenger
display_name: 挑戰者
model: google/gemini-2.5-pro
stance: critique
temperature: 0.5
color: "#E67E22"
focus_areas:
  - 假設挑戰
  - 論點弱點
  - 替代觀點
system_prompt: |
  你是一位專業的文學挑戰者。你的任務是：
  1. 找出文本中未被質疑的預設立場或假設
  2. 挑戰敘事邏輯中的弱點或跳躍
  3. 提出作者可能沒想到的替代觀點或可能性
  4. 對其他 Agent 過於寬容的評語提出反駁

  你的價值在於「讓作者不舒服但有收穫」。
```

**內建 Preset（3 組）**

```yaml
# presets.yaml
- name: editorial
  display_name: 編輯室
  description: 標準文章審議 — 編輯、讀者、結構三方意見
  agents: [strict-editor, casual-reader, structure-analyst]

- name: debate
  display_name: 辯論場
  description: 正反對抗 — 辯護者 vs 挑戰者，嚴格編輯裁決
  agents: [advocate, challenger, strict-editor]

- name: publish-ready
  display_name: 出版前檢查
  description: 全方位檢查 — 編輯、結構、行銷、讀者
  agents: [strict-editor, structure-analyst, marketing-eye, casual-reader]
```

### 7.4 流水線 Orchestrator（狀態機）

**狀態定義**

```go
// orchestrator/state.go

type PipelineState int
const (
    StateIdle         PipelineState = iota  // 初始 / 等待啟動
    StateInitializing                       // 載入 agents、稿件、設定
    StateReady                              // 已就緒，等使用者指令
    StateAgentTurn                          // 某 Agent 正在發言（串流中）
    StateRoundEnd                           // 一輪結束，等使用者介入
    StateCompressing                        // Context 壓縮中
    StateSummarizing                        // Summarizer Agent 執行中
    StateCompleted                          // 審議完成
    StatePaused                             // 使用者暫停
    StateError                              // 不可恢復錯誤
)
```

**狀態遷移圖**

```
                ┌──────────────────────────────────────────┐
                │                                          │
                ▼                                          │
  [Idle] ──▶ [Initializing] ──▶ [Ready]                   │
                                   │                       │
                          使用者啟動 / /start               │
                                   │                       │
                                   ▼                       │
                  ┌──────── [AgentTurn] ◄────────┐        │
                  │             │                 │        │
                  │      串流完成/Agent done       │        │
                  │             │                 │        │
                  │             ▼                 │        │
                  │    {還有下一個 Agent?}         │        │
                  │     Yes ──▶ [AgentTurn]       │        │
                  │     No  ──▶ [RoundEnd]        │        │
                  │                 │              │        │
                  │        ┌───────┴────────┐     │        │
                  │        ▼                ▼     │        │
                  │  使用者介入         /next     │        │
                  │   (文字/指令)        │        │        │
                  │        │             │        │        │
                  │        ▼             ▼        │        │
                  │  {需要壓縮 context?}          │        │
                  │   Yes ──▶ [Compressing]──┐    │        │
                  │   No  ────────────────┐  │    │        │
                  │                       ▼  ▼    │        │
                  │                  [AgentTurn] ──┘        │
                  │                                        │
                  │  使用者 /end 或 達到 max_rounds         │
                  │        │                               │
                  │        ▼                               │
                  │  {啟用 Summarizer?}                     │
                  │   Yes ──▶ [Summarizing] ──▶ [Completed]│
                  │   No  ──────────────────▶ [Completed]  │
                  │                                        │
                  │  使用者 /pause                          │
                  └─────────▶ [Paused] ────resume──────────┘

  任何狀態 ── 不可恢復錯誤 ──▶ [Error]
```

**Orchestrator Interface**

```go
// orchestrator/pipeline.go

type Pipeline interface {
    // Initialize 載入 session 設定、agents、稿件
    Initialize(ctx context.Context, cfg PipelineConfig) error

    // Start 啟動流水線
    Start(ctx context.Context) error

    // NextRound 執行下一輪
    NextRound(ctx context.Context) error

    // Intervene 使用者介入
    Intervene(ctx context.Context, input Intervention) error

    // Pause 暫停
    Pause() error

    // Resume 從暫停恢復
    Resume(ctx context.Context) error

    // End 結束審議
    End(ctx context.Context) error

    // State 查詢目前狀態
    State() PipelineState

    // Subscribe 訂閱事件（供 TUI 監聽）
    Subscribe() <-chan PipelineEvent
}

type PipelineConfig struct {
    Session     *session.Session
    Agents      []agents.Role
    MaxRounds   int
    Summarizer  *agents.Role   // nil = 不啟用
    BudgetLimit float64        // 0 = 無限制
}

type Intervention struct {
    Type    InterventionType
    Content string             // 使用者輸入
    Focus   *LineRange         // /focus 指令用
}

type InterventionType int
const (
    InterventionChat     InterventionType = iota  // 一般追問
    InterventionFocus                              // 聚焦特定段落
    InterventionAddFile                            // 補充檔案
    InterventionSkip                               // 跳過某 Agent
)

type PipelineEvent struct {
    Type      EventType
    AgentName string
    Delta     string            // 串流增量
    Round     int
    State     PipelineState
    Usage     *openrouter.Usage
    Error     error
}
```

**Context 組裝邏輯（虛擬碼）**

```
function buildContext(agent, round, session):
    messages = []

    // L1: 系統 Prompt
    messages.append(system: agent.systemPrompt)

    // L2: 原稿
    messages.append(user: formatDraft(session.source))

    // L3: 使用者開場 + 介入指令
    for intervention in session.interventions:
        messages.append(user: intervention.content)

    // L5: 舊輪次摘要（若有）
    for oldRound in session.rounds where oldRound < (round - 2):
        if session.summaries[oldRound] exists:
            messages.append(user: "[前序討論摘要]\n" + summary)

    // L4: 最近 2 輪完整內容
    for recentRound in [round-2, round-1]:
        for response in session.rounds[recentRound].responses:
            messages.append(assistant: formatResponse(response))

    // 當前輪次已有的回應
    for priorResponse in currentRound.responses:
        messages.append(assistant: formatResponse(priorResponse))

    // 估算 token，若超過 70% context window → 觸發壓縮
    if estimateTokens(messages) > agent.model.contextLen * 0.70:
        trigger compression

    return messages
```

### 7.5 Context 管理

**分層保留策略**

| 優先級 | 內容 | 處理方式 | 可犧牲 |
|--------|------|----------|--------|
| L1 | 系統 Prompt | 永遠保留 | ❌ |
| L2 | 原始稿件 | 永遠保留 | ❌ |
| L3 | 使用者介入指令 | 永遠保留 | ❌ |
| L4 | 最近 2 輪完整討論 | 完整保留 | ❌ |
| L5 | 較舊輪次 | 壓縮為摘要（每輪 ≤ 200 token） | ✅（壓縮） |
| L6 | 最舊內容 | 必要時捨棄 | ✅（捨棄） |

**壓縮觸發條件**
- 下一次 API 呼叫預估 token 超過模型 context window 的 70%
- 觸發後，由 `config.context.summary_model` 指定的輕量模型壓縮

**壓縮 Prompt 模板**

```
你是一個審議討論的摘要員。請將以下一輪討論壓縮為簡潔摘要（≤200字），
保留：每位 Agent 的核心觀點、立場標籤、關鍵分歧。
刪除：修辭、重複論點、客套話。

[以下為第 {round} 輪討論原文]
{round_content}
```

**Token 估算方式**
- 使用 tiktoken-go（或等效）對 message 內容估算 token 數
- 估算值加 10% 安全邊距
- 若無法使用 tiktoken，以「1 token ≈ 4 字元（英文）/ 1.5 字元（中文）」作為粗估

### 7.6 Ingestion 模組

**統一介面**

```go
// ingestion/ingester.go

type Ingester interface {
    // Ingest 讀取檔案並轉為統一格式
    Ingest(ctx context.Context, path string) (*Document, error)

    // SupportedExts 回傳支援的副檔名
    SupportedExts() []string
}

type Document struct {
    SourcePath  string
    SourceType  string        // "md" | "txt" | "pdf"
    Content     string        // Markdown 格式的內容
    Metadata    DocMetadata
    LineCount   int
    TokenEstimate int
}

type DocMetadata struct {
    Title       string        // 若可推斷
    WordCount   int
    CharCount   int
    ContentHash string        // SHA-256
}
```

**PDF 解析器 — 技術規格（自建）**

Phase 1 範圍定義：

| 項目 | 支援 | 不支援 |
|------|------|--------|
| 文字層 PDF | ✅ | — |
| 掃描件 / 影像 PDF | — | ❌ |
| Standard 14 字型 | ✅ | — |
| TrueType / Type1 嵌入字型 | ✅（基本） | 複雜 subsetting |
| ToUnicode CMap | ✅ | — |
| Predefined CMap（CJK） | ✅（常見） | 罕見 CMap |
| FlateDecode 壓縮 | ✅ | — |
| LZWDecode | ❌（Phase 2） | — |
| 單欄佈局 | ✅ | — |
| 多欄佈局 | ❌（Phase 2） | — |
| 表格 | ❌（Phase 2） | — |
| 圖片 | ❌ | — |
| 加密 PDF | ❌ | — |
| PDF/A | ✅（子集） | — |

**PDF 解析流程**

```
檔案讀取
    │
    ▼
[lexer.go] Token 化
    PDF 操作符、字串、數字、名稱、字典、陣列
    │
    ▼
[xref.go] 解析 Cross-Reference Table
    建立物件位址索引
    │
    ▼
[parser.go] 建構物件樹
    Catalog → Pages → Page 逐層解析
    │
    ▼
[stream.go] Stream 解碼
    FlateDecode (zlib) 解壓縮
    │
    ▼
[extractor.go] 文字提取
    走訪每頁 content stream
    解析 BT/ET 文字物件、Tj/TJ 操作符
    收集 (x, y, text) 三元組
    │
    ▼
[encoder.go + cmap.go] 編碼轉換
    字元碼 → Unicode
    處理 ToUnicode CMap / Encoding 字典
    │
    ▼
[markdown.go] 組合 Markdown
    依 y 座標分行
    依行間距分段
    輸出 Markdown
```

**PDF 模組 Interface**

```go
// ingestion/pdf/

type PDFParser interface {
    Parse(r io.ReadSeeker) (*PDFDocument, error)
}

type PDFDocument struct {
    Pages    []Page
    Metadata map[string]string
}

type Page struct {
    Number   int
    TextRuns []TextRun
}

type TextRun struct {
    X, Y     float64
    FontSize float64
    FontName string
    Text     string
}

type TextExtractor interface {
    Extract(doc *PDFDocument) string   // 輸出 Markdown
}
```

### 7.7 Session 管理

**儲存路徑結構**

```
~/.local/share/woolf/
├── sessions/
│   ├── 20260507-143022-chapter3.json
│   ├── 20260506-091500-opening.json
│   └── ...
├── agents/                # 使用者自訂角色
│   └── my-custom-role.yaml
└── cache/
    └── models.json        # 模型清單快取（TTL: 24h）
```

**Session 生命週期**

```
created → active → paused → active → completed
                                   → abandoned（使用者刪除）
                         → error（不可恢復）
```

**Session Interface**

```go
// session/store.go

type Store interface {
    Create(cfg SessionConfig) (*Session, error)
    Load(id string) (*Session, error)
    Save(s *Session) error
    List(filter ListFilter) ([]SessionSummary, error)
    Delete(id string) error
    Fork(id string, opts ForkOptions) (*Session, error)
    Search(query string) ([]SessionSummary, error)
}

type ListFilter struct {
    Limit  int
    Since  *time.Time
    Status *SessionStatus
}

type SessionSummary struct {
    ID        string
    Title     string
    Status    SessionStatus
    CreatedAt time.Time
    UpdatedAt time.Time
    Rounds    int
    TotalCost float64
}

type ForkOptions struct {
    NewDraftPath string
    NewTitle     string
}
```

**自動儲存時機**
- 每個 Agent 回應完成後
- 使用者介入後
- 每輪結束後
- 程式異常退出前（graceful shutdown hook）

### 7.8 匯出模組

**Markdown 匯出格式**

```markdown
# Woolf 審議記錄

**Session**: 20260507-143022-chapter3
**日期**: 2026-05-07 14:30 ~ 15:12
**Agent 組合**: 嚴格編輯 (GPT-4o) / 一般讀者 (Claude Sonnet) / 結構分析師 (Gemini Pro)

---

## 原稿

> [稿件內容，以 blockquote 呈現]

---

## Round 1

### 🔴 嚴格編輯 (GPT-4o)

[回應內容]

### 🔵 一般讀者 (Claude Sonnet) → 回應：嚴格編輯 [disagree]

[回應內容]

### 🟢 結構分析師 (Gemini Pro) → 回應：嚴格編輯 [agree], 一般讀者 [extend]

[回應內容]

---

## Round 2

[...]

---

## 統計

| 項目 | 數值 |
|------|------|
| 總輪數 | 3 |
| 總 Token | 12,450 |
| 總費用 | $0.089 |
```

**PDF 匯出**（Phase 2，暫定方向）
- 優先評估使用 Typst 作為排版引擎
- 若 Typst 不可行，退回 Go 原生 PDF 產生（如 gofpdf）
- 排版風格：書籍審閱報告形式

### 7.9 成本追蹤

```go
// cost/tracker.go

type Tracker interface {
    // Record 記錄一次 API 呼叫的消耗
    Record(entry UsageEntry)

    // SessionTotal 取得 session 累計
    SessionTotal() CostSummary

    // CheckBudget 檢查是否超過預算
    CheckBudget(limit float64) BudgetStatus
}

type UsageEntry struct {
    Model            string
    PromptTokens     int
    CompletionTokens int
    Timestamp        time.Time
}

type CostSummary struct {
    TotalTokens      int
    TotalPrompt      int
    TotalCompletion  int
    TotalCostUSD     float64
    ByModel          map[string]ModelCost
}

type BudgetStatus struct {
    Used       float64
    Limit      float64
    Percentage float64    // 0.0 ~ 1.0
    Exceeded   bool
    Warning    bool       // 超過 warn_threshold
}
```

**定價資料來源**
- 首次啟動時從 OpenRouter API 拉取模型定價
- 快取於 `~/.local/share/woolf/cache/models.json`，TTL 24 小時
- 若快取失效且網路不可用，使用上次快取值並標記「定價可能不準確」

### 7.10 Configuration

**設定檔路徑**（遵循 XDG）

| 平台 | 路徑 |
|------|------|
| Linux | `~/.config/woolf/config.toml` |
| macOS | `~/Library/Application Support/woolf/config.toml` |
| Windows | `%APPDATA%\woolf\config.toml` |

**完整設定檔結構**

```toml
# Woolf Configuration

[api]
openrouter_key = ""                    # 或使用環境變數 OPENROUTER_API_KEY
base_url = "https://openrouter.ai/api/v1"
timeout_seconds = 120
max_retries = 3

[defaults]
max_rounds = 3
auto_save = true
language = "zh-TW"
default_preset = "editorial"
summarizer_enabled = false
summarizer_model = "openai/gpt-4o-mini"

[tui]
theme = "dark"                         # "dark" | "light"
show_token_count = true
show_cost = true
stream_speed = "realtime"              # "realtime" | "fast"（跳過串流動畫）
editor = "vim"                         # 用於 woolf config edit

[context]
max_window_ratio = 0.70
summary_model = "openai/gpt-4o-mini"
summary_max_tokens = 200

[budget]
session_limit_usd = 0.00              # 0 = 無限制
warn_threshold = 0.80

[paths]
sessions_dir = ""                      # 空 = 使用 XDG 預設
agents_dir = ""                        # 空 = 使用 XDG 預設

# Agent 定義可在此內聯，或放在獨立 YAML 檔案
# 獨立檔案放在 agents_dir 中，會自動載入
```

**設定載入優先級**（高 → 低）
1. CLI flags（`--config`, `--verbose` 等）
2. 環境變數（`OPENROUTER_API_KEY`, `WOOLF_CONFIG` 等）
3. 設定檔（config.toml）
4. 內建預設值

**環境變數對應**

| 環境變數 | 對應設定 |
|----------|----------|
| `OPENROUTER_API_KEY` | `api.openrouter_key` |
| `WOOLF_CONFIG` | 設定檔路徑覆蓋 |
| `WOOLF_SESSIONS_DIR` | `paths.sessions_dir` |
| `WOOLF_DEBUG` | 啟用 debug 模式 |
| `NO_COLOR` | 停用顏色（標準慣例） |

### 7.11 TUI 規格

**框架**：[Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles)

**佈局結構**

```
┌─ Woolf ─ Session: {title} ─ Round {n}/{max} ─ {status} ──────────────────┐
│                                                                           │
│  ╭─ 討論串 ─────────────────────────────────────────────────────────╮    │
│  │                                                                   │    │
│  │  ┌──────────────────────────────────────────────────────────┐    │    │
│  │  │ 🔴 嚴格編輯 (GPT-4o) · Round 1                          │    │    │
│  │  │ 第二段的轉折太突兀，「她轉身離開」這個動作缺乏前置情     │    │    │
│  │  │ 緒鋪墊，建議加入一段內心猶豫的描寫...                    │    │    │
│  │  └──────────────────────────────────────────────────────────┘    │    │
│  │                                                                   │    │
│  │  ┌──────────────────────────────────────────────────────────┐    │    │
│  │  │ 🔵 一般讀者 (Claude) · Round 1 → 回應嚴格編輯 [disagree]│    │    │
│  │  │ 我反而覺得「突然轉身」的突兀感是這段最有張力的地方。     │    │    │
│  │  │ 如果加太多內心戲，反而會拖慢節奏...                      │    │    │
│  │  └──────────────────────────────────────────────────────────┘    │    │
│  │                                                                   │    │
│  │  ┌──────────────────────────────────────────────────────────┐    │    │
│  │  │ 🟢 結構分析師 (Gemini) · Round 1  (串流中... ▊)         │    │    │
│  │  │ 從敘事弧線的角度來看...                                  │    │    │
│  │  └──────────────────────────────────────────────────────────┘    │    │
│  │                                                                   │    │
│  ╰───────────────────────────────────────────────────────────────────╯    │
│                                                                           │
│  ╭─ 輸入 ───────────────────────────────────────────────────────────╮    │
│  │ > _                                                               │    │
│  │                                                                   │    │
│  │ [Tab: 切換面板] [/: 指令] [Ctrl+S: 匯出] [Esc: 暫停]            │    │
│  ╰───────────────────────────────────────────────────────────────────╯    │
│                                                                           │
│  ╭─ 狀態列 ─────────────────────────────────────────────────────────╮    │
│  │ Tokens: 4,521 / 128K │ Cost: $0.023 / $0.50 │ R: 1/3 │ ● Live  │    │
│  ╰───────────────────────────────────────────────────────────────────╯    │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

**鍵盤快捷鍵**

| 按鍵 | Context | 功能 |
|------|---------|------|
| `Tab` | 全域 | 切換焦點面板（討論串 ↔ 輸入區） |
| `j` / `↓` | 討論串 | 向下捲動 |
| `k` / `↑` | 討論串 | 向上捲動 |
| `g` | 討論串 | 跳到最頂 |
| `G` | 討論串 | 跳到最底 |
| `Enter` | 輸入區 | 送出輸入 / 繼續下一輪 |
| `Shift+Enter` | 輸入區 | 換行（多行輸入） |
| `/` | 輸入區 | 進入指令模式 |
| `Ctrl+S` | 全域 | 手動觸發匯出 |
| `Ctrl+C` | 全域 | 若在串流中：中斷當前 Agent；否則：結束程式 |
| `Esc` | 全域 | 暫停審議 |
| `?` | 全域 | 顯示快捷鍵幫助 |

**TUI 指令模式（`/` 開頭）**

| 指令 | 功能 |
|------|------|
| `/start` | 啟動流水線 |
| `/next` | 繼續下一輪 |
| `/end` | 結束審議 |
| `/pause` | 暫停 |
| `/focus <line_start>-<line_end>` | 聚焦特定行 |
| `/add-file <path>` | 補充檔案 |
| `/skip <agent_name>` | 下一輪跳過某 Agent |
| `/summarize` | 觸發摘要 |
| `/export md` / `/export pdf` | 匯出 |
| `/agents` | 顯示當前 Agent 清單 |
| `/status` | 顯示詳細狀態 |
| `/help` | 指令說明 |
| `/quit` | 儲存並退出 |

**色彩主題系統**

```go
// tui/theme.go

type Theme struct {
    Name           string
    Background     lipgloss.Color
    Foreground     lipgloss.Color
    BorderColor    lipgloss.Color
    HeaderBG       lipgloss.Color
    StatusBG       lipgloss.Color
    AgentColors    map[string]lipgloss.Color
    StanceTags     map[Stance]lipgloss.Color
    HighlightColor lipgloss.Color
    DimColor       lipgloss.Color
    ErrorColor     lipgloss.Color
    WarningColor   lipgloss.Color
}

var DarkTheme = Theme{
    Name:           "dark",
    Background:     lipgloss.Color("#1E1E2E"),
    Foreground:     lipgloss.Color("#CDD6F4"),
    BorderColor:    lipgloss.Color("#6C7086"),
    HeaderBG:       lipgloss.Color("#313244"),
    StatusBG:       lipgloss.Color("#181825"),
    StanceTags: map[Stance]lipgloss.Color{
        "agree":    lipgloss.Color("#A6E3A1"),
        "disagree": lipgloss.Color("#F38BA8"),
        "extend":   lipgloss.Color("#89DCEB"),
        "neutral":  lipgloss.Color("#9399B2"),
    },
    HighlightColor: lipgloss.Color("#F5C2E7"),
    DimColor:       lipgloss.Color("#585B70"),
    ErrorColor:     lipgloss.Color("#F38BA8"),
    WarningColor:   lipgloss.Color("#FAB387"),
}

// Light theme 從 DarkTheme 反轉
```

**串流渲染規則**
- 每接收到一個 `StreamEvent.Delta`，立即 append 到當前 Agent 的顯示區
- 更新頻率：每個 delta 即觸發 TUI 重繪
- 自動捲動：串流中自動捲到底部
- 串流結束後：顯示 Agent badge + 立場標籤 + token 統計

---

## 8. 資料格式定義

### 8.1 Session JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Woolf Session",
  "type": "object",
  "required": ["session_id", "version", "status", "created_at", "agents_config", "rounds"],
  "properties": {
    "session_id": {
      "type": "string",
      "pattern": "^\\d{8}-\\d{6}-.+$"
    },
    "version": {
      "type": "string",
      "const": "1.0"
    },
    "title": { "type": "string" },
    "status": {
      "type": "string",
      "enum": ["active", "paused", "completed", "error"]
    },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "source": {
      "type": "object",
      "properties": {
        "type": { "enum": ["file", "input"] },
        "path": { "type": "string" },
        "content": { "type": "string" },
        "content_hash": { "type": "string" },
        "content_preview": { "type": "string", "maxLength": 200 }
      }
    },
    "agents_config": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "display_name": { "type": "string" },
          "model": { "type": "string" },
          "stance": { "enum": ["critique", "support", "neutral"] },
          "order": { "type": "integer" },
          "color": { "type": "string" }
        }
      }
    },
    "rounds": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "round_index": { "type": "integer" },
          "started_at": { "type": "string", "format": "date-time" },
          "completed_at": { "type": "string", "format": "date-time" },
          "responses": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "agent_name": { "type": "string" },
                "model": { "type": "string" },
                "responding_to": {
                  "type": ["string", "null"]
                },
                "stance_tag": {
                  "type": ["string", "null"],
                  "enum": ["agree", "disagree", "extend", "neutral", null]
                },
                "content": { "type": "string" },
                "tokens": {
                  "type": "object",
                  "properties": {
                    "prompt": { "type": "integer" },
                    "completion": { "type": "integer" }
                  }
                },
                "cost_usd": { "type": "number" },
                "timestamp": { "type": "string", "format": "date-time" },
                "status": {
                  "enum": ["completed", "interrupted", "skipped", "error"]
                }
              }
            }
          }
        }
      }
    },
    "interventions": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "after_round": { "type": "integer" },
          "type": { "enum": ["chat", "focus", "add_file", "skip"] },
          "content": { "type": "string" },
          "focus_range": {
            "type": ["object", "null"],
            "properties": {
              "start_line": { "type": "integer" },
              "end_line": { "type": "integer" }
            }
          },
          "timestamp": { "type": "string", "format": "date-time" }
        }
      }
    },
    "summaries": {
      "type": "object",
      "additionalProperties": { "type": "string" }
    },
    "totals": {
      "type": "object",
      "properties": {
        "rounds_completed": { "type": "integer" },
        "total_tokens": { "type": "integer" },
        "total_prompt_tokens": { "type": "integer" },
        "total_completion_tokens": { "type": "integer" },
        "total_cost_usd": { "type": "number" }
      }
    }
  }
}
```

---

## 9. 錯誤處理策略

### 9.1 錯誤分類

| 類別 | 代碼前綴 | 範例 | 處理策略 |
|------|----------|------|----------|
| **配置錯誤** | `CFG-` | API key 未設定、設定檔格式錯誤 | 啟動時攔截，顯示修復指引 |
| **網路錯誤** | `NET-` | 超時、DNS 失敗、連線中斷 | 重試 → 通知 TUI → 使用者決定 |
| **API 錯誤** | `API-` | 429 Rate Limit、401 Unauthorized、模型下線 | 依類型處理（見下方） |
| **Ingestion 錯誤** | `ING-` | 檔案不存在、PDF 解析失敗、編碼錯誤 | 通知使用者，提供替代方案 |
| **Session 錯誤** | `SES-` | 儲存失敗、磁碟空間不足、JSON 損毀 | 嘗試備份 → 通知使用者 |
| **TUI 錯誤** | `TUI-` | 終端機不支援、尺寸太小 | 降級顯示或退出並提示 |

### 9.2 API 錯誤處理細節

| HTTP Status | 錯誤碼 | 處理 |
|-------------|--------|------|
| 401 | `API-001` | 停止流水線，提示重新設定 API key |
| 402 | `API-002` | 停止流水線，提示 OpenRouter 餘額不足 |
| 429 | `API-003` | 讀取 Retry-After，等待後重試（最多 3 次） |
| 模型 404 | `API-004` | 標記該 Agent 為 skipped，流水線繼續；TUI 提示 |
| 500-503 | `API-005` | 指數退避重試（1s, 2s, 4s），最多 3 次 |

### 9.3 Graceful Degradation

| 狀況 | 行為 |
|------|------|
| 單一 Agent 失敗 | 標記為 `skipped`，其他 Agent 繼續 |
| 所有 Agent 失敗 | 暫停流水線，提示使用者檢查網路 / API |
| Session 儲存失敗 | 嘗試寫入備份路徑（`/tmp/woolf-backup-{id}.json`） |
| PDF 解析部分失敗 | 輸出已解析部分 + 警告標記 `[⚠ 此處解析失敗]` |
| 終端機尺寸 < 80×24 | 顯示警告，建議調整視窗大小 |

### 9.4 錯誤日誌

- 路徑：`~/.local/share/woolf/logs/woolf-{date}.log`
- 格式：結構化 JSON 日誌
- 層級：DEBUG / INFO / WARN / ERROR
- 預設：僅記錄 WARN 以上；`--debug` 啟用全部
- 輪替：保留最近 7 天

---

## 10. 安全性考量

| 項目 | 風險 | 緩解措施 |
|------|------|----------|
| **API Key 儲存** | 明文存在設定檔 | 設定檔權限 `0600`；建議使用環境變數；`woolf config show` 遮蔽 key 顯示 |
| **Session 中的稿件** | 本地明文儲存 | Session 目錄權限 `0700`；提醒使用者注意共用電腦風險 |
| **API 傳輸** | 稿件透過 HTTPS 傳送至 OpenRouter | 僅使用 HTTPS；提醒使用者敏感內容風險 |
| **依賴供應鏈** | Go module 被篡改 | 使用 `go.sum` 校驗；定期 `govulncheck` |
| **日誌洩漏** | 日誌可能包含稿件片段 | DEBUG 層級日誌不記錄完整 API 回應；WARN 以上不含稿件內容 |

**安全預設**
- `woolf init` 建立的設定檔自動設定 `0600` 權限
- Session 目錄自動設定 `0700` 權限
- API key 在 TUI 和日誌中永遠以 `sk-or-...****` 遮蔽顯示

---

## 11. 測試策略

### 11.1 測試層級

| 層級 | 範圍 | 工具 | 覆蓋目標 |
|------|------|------|----------|
| **單元測試** | 各模組內部邏輯 | Go testing + testify | ≥ 70% |
| **整合測試** | 模組間互動 | Go testing + mock | 關鍵路徑 100% |
| **端對端測試** | CLI 指令 → 完整流程 | Go testing + exec | 核心 flow |
| **PDF 回歸測試** | PDF 解析器 | 固定測試集 | 見下方 |

### 11.2 Mock 策略

| 模組 | Mock 方式 |
|------|-----------|
| `openrouter` | Interface mock：回傳預定義 StreamEvent |
| `session/store` | In-memory store 實作 |
| `ingestion/pdf` | 使用固定測試 PDF 檔案 |
| `cost/pricing` | 硬編碼定價資料 |

### 11.3 PDF 解析器測試集

```
testdata/pdf/
├── simple-text.pdf          # 純文字，單欄
├── chinese-text.pdf         # 中文內容
├── mixed-encoding.pdf       # 多種編碼混合
├── standard14-fonts.pdf     # Standard 14 字型
├── embedded-font.pdf        # 嵌入 TrueType
├── multipage.pdf            # 多頁
├── compressed.pdf           # FlateDecode 壓縮
├── unsupported-encrypted.pdf   # 加密（應回傳錯誤）
├── unsupported-scan.pdf        # 掃描件（應回傳錯誤）
└── expected/                   # 預期輸出
    ├── simple-text.md
    ├── chinese-text.md
    └── ...
```

每份測試 PDF 附帶對應的預期 Markdown 輸出。CI 中比對實際輸出與預期輸出，差異超過閾值則失敗。

### 11.4 CI Pipeline

```
GitHub Actions:
  - on: [push, pull_request]
  - jobs:
    - lint (golangci-lint)
    - test (go test -race -cover ./...)
    - build (cross-compile: linux/amd64, darwin/amd64, darwin/arm64, windows/amd64)
    - pdf-regression (獨立 job，跑 PDF 測試集)
```

---

## 12. 安裝與發佈

### 12.1 安裝方式

| 方式 | 指令 |
|------|------|
| **Go install** | `go install github.com/<org>/woolf/cmd/woolf@latest` |
| **Homebrew** (macOS/Linux) | `brew install woolf` |
| **GitHub Release** | 下載預編譯二進位檔 |
| **手動編譯** | `git clone ... && make build` |

### 12.2 Build Matrix

| OS | Arch | 備註 |
|----|------|------|
| Linux | amd64 | 主要目標 |
| Linux | arm64 | 支援 |
| macOS | amd64 | 支援 |
| macOS | arm64 (Apple Silicon) | 主要目標 |
| Windows | amd64 | Windows Terminal 建議 |

### 12.3 發佈流程

```
tag vX.Y.Z → GitHub Actions →
  build all targets →
  create GitHub Release →
  upload binaries →
  update Homebrew formula（if applicable）
```

### 12.4 版本號規則（Semantic Versioning）
- `0.x.y`：Pre-release，API 不穩定
- `1.0.0`：Phase 1 功能完成，Session 格式穩定
- `MAJOR`：Session 格式不相容變更
- `MINOR`：新功能
- `PATCH`：Bug fix

---

## 13. 風險與緩解

| # | 風險 | 嚴重度 | 機率 | 緩解策略 |
|---|------|--------|------|----------|
| R1 | PDF 自建解析器工程量超出預期 | 🔴 高 | 🟠 中 | MVP 嚴格限縮；獨立子專案先行 spike；Ingester interface 設計為可替換，必要時可插入現成 lib |
| R2 | PDF 匯出排版複雜度 | 🔴 高 | 🟠 中 | Phase 1 僅支援 .md 匯出；Phase 2 評估 Typst / LaTeX 方案 |
| R3 | Context 膨脹導致回應品質下降 | 🟠 中 | 🟠 中 | 分層保留 + 漸進摘要（§7.5）；使用者可手動 `/focus` 縮窄範圍 |
| R4 | OpenRouter 模型頻繁異動 | 🟠 中 | 🟡 低 | 動態查詢模型清單；角色定義支援 fallback model |
| R5 | 多 Agent 討論離題或品質不穩 | 🟠 中 | 🟠 中 | System prompt 嚴格錨定；temperature 調低；提供 `/focus` 機制 |
| R6 | API 費用超出使用者預期 | 🟡 低 | 🟠 中 | 即時成本顯示；預算上限 + 警告機制 |
| R7 | TUI 跨終端機相容問題 | 🟡 低 | 🟡 低 | 依賴 Bubble Tea 成熟的跨平台支援；提供 `--no-color` fallback |
| R8 | Session 資料遺失 | 🟡 低 | 🟡 低 | 每步自動儲存；graceful shutdown hook；備份路徑機制 |
| R9 | 單一 Agent 回應過長佔滿畫面 | 🟡 低 | 🟠 中 | 設定 `max_tokens` 上限；TUI 支援摺疊 / 展開 |

---

## 14. KPI 與驗收標準

### 14.1 Phase 1 驗收矩陣

| 領域 | 項目 | 驗收標準 | 優先級 |
|------|------|----------|--------|
| **流水線** | 基本審議 | 3 Agent × 3 輪完整執行，context 正確傳遞 | Must |
| **流水線** | 立場標籤 | 每位 Agent 正確標記 stance，TUI 正確顯示 | Must |
| **流水線** | 使用者介入 | 介入內容被納入下一輪 context | Must |
| **流水線** | 中斷續接 | Ctrl+C / `/pause` 後可 `resume` | Must |
| **TUI** | 串流渲染 | 字元逐步顯示，延遲 < 100ms | Must |
| **TUI** | 長文輸入 | 5000 字輸入不崩潰不卡頓 | Must |
| **TUI** | 鍵盤操作 | 所有功能可純鍵盤操作 | Must |
| **TUI** | 指令模式 | `/` 指令正確執行 | Must |
| **檔案** | .md 讀取 | 中英混排無亂碼 | Must |
| **檔案** | .txt 讀取 | UTF-8 無亂碼 | Must |
| **檔案** | .pdf 讀取 | Phase 1 範圍 PDF 提取準確率 ≥ 85% | Must |
| **Session** | 自動儲存 | 每步完成後 5 秒內寫入磁碟 | Must |
| **Session** | 列表瀏覽 | `woolf list` 正確顯示歷史 | Must |
| **匯出** | .md 匯出 | 結構完整、Agent 標識清楚 | Must |
| **成本** | 計費精準度 | 與 OpenRouter Dashboard 誤差 < 5% | Must |
| **設定** | 自訂 Agent | YAML 載入後可立即使用 | Must |
| **Session** | TUI 瀏覽器 | TUI 內互動式瀏覽歷史 | Should |
| **Context** | 自動摘要 | 超過閾值時正確觸發壓縮 | Should |
| **匯出** | .pdf 匯出 | 排版合理、可閱讀 | Should |
| **Session** | 分叉 | fork 後可獨立執行 | Should |

### 14.2 效能指標

| 指標 | 目標 |
|------|------|
| CLI 冷啟動時間 | < 500ms |
| TUI 渲染 FPS | ≥ 30 FPS（串流中） |
| Session 儲存延遲 | < 1 秒（一般大小） |
| PDF 解析速度 | < 5 秒（50 頁以內） |
| 記憶體使用 | < 200 MB（一般 session） |

---

## 15. Phase 2 路線圖

### 15.1 Phase 2 功能清單

| 優先序 | 功能 | 預估工作量 | 依賴 |
|--------|------|------------|------|
| P2-1 | Summarizer Agent | 1 週 | Phase 1 完成 |
| P2-2 | 沙龍記憶（跨 Session Agent 個性） | 2 週 | Session + Agent 系統 |
| P2-3 | 辯論模式（Agent 對戰） | 1.5 週 | Orchestrator 擴展 |
| P2-4 | Micro-session（段落審議） | 1 週 | TUI + Orchestrator |
| P2-5 | Session diff 比較 | 1.5 週 | Session fork |
| P2-6 | 分歧度指標 | 1 週 | Summarizer |
| P2-7 | 重播模式 | 1 週 | Session 時序資料 |
| P2-8 | PDF 多欄偵測 | 2 週 | PDF Phase 1 |
| P2-9 | PDF 匯出（Typst） | 2 週 | 獨立 |
| P2-10 | 本地模型支援（Ollama） | 1.5 週 | OpenRouter Client interface 擴展 |
| P2-11 | 插件系統 | 3 週 | 整體架構穩定後 |

### 15.2 Phase 2 時程（暫估）

```
Month 1: P2-1 Summarizer + P2-4 Micro-session
Month 2: P2-2 沙龍記憶 + P2-3 辯論模式
Month 3: P2-5 Session diff + P2-6 分歧度 + P2-7 重播
Month 4: P2-8 PDF 多欄 + P2-9 PDF 匯出
Month 5: P2-10 Ollama + P2-11 插件系統
```

---

## 16. 創意延伸

### 16.1 「沙龍記憶」— 跨 Session 的 Agent 個性沉澱
**價值**：每位 Agent 在多次審議中累積對作者風格的認識，逐漸發展出對該作者更精準的回饋。從「冷啟動」進化為「越用越懂你」。

**實作方式**：
- 每個 Agent 在 session 結束時，由輕量模型產出一份「對作者觀察筆記」（≤ 500 token）
- 儲存於 `~/.local/share/woolf/memory/{agent_name}.json`
- 下次同 Agent 啟動時，將觀察筆記注入 system prompt 尾端
- 支援使用者清除特定 Agent 的記憶（`woolf agents reset-memory <name>`）

### 16.2 「逐句質詢」模式（Micro-session）
**價值**：對特定段落取得多元意見，提高審議效率，避免 Agent 每次都從頭評論。

**實作方式**：
- TUI 支援 `/focus 12-18` 選取稿件第 12-18 行
- 啟動 micro-session，context 僅含選取段落 + 上下文各 5 行
- 回應合併回主 session 的 `micro_annotations` 欄位
- micro-session 不計入主輪次，但費用統一計算

### 16.3 「分歧度指標」
**價值**：視覺化呈現 Agent 間共識 / 爭議程度，幫助作者快速定位最需要關注的問題。

**實作方式**：
- 每輪結束後，由輕量模型分析所有回應，產出：
  ```json
  {
    "topics": [
      {
        "topic": "第二段轉折",
        "positions": {
          "strict-editor": "needs_fix",
          "casual-reader": "acceptable",
          "structure-analyst": "needs_fix"
        },
        "divergence": 0.33
      }
    ]
  }
  ```
- TUI 狀態列以 ASCII bar 顯示：`Consensus: ████░░ 67%`
- 高分歧議題標記醒目色

### 16.4 「重播模式」
**價值**：完成的 session 可逐字串流重播，用於教學、團隊分享、或寫作回顧。

**實作方式**：
- `woolf replay <session-id> [--speed 2x]`
- 依原始時間戳差值模擬串流速度
- 支援暫停、快進、快退
- 唯讀模式，不可修改

### 16.5 「角色對戰」辯論模式
**價值**：讓兩位 Agent 就特定議題正面交鋒，逼出更深刻的論點。

**實作方式**：
- `woolf start --mode debate --agents advocate,challenger --judge strict-editor`
- 辯論流程：Agent A → Agent B 回擊 → Agent A 再回擊 → ... → Judge 裁決
- 裁決輸出包含：各方優劣分析、最終建議、爭議未決項目
- 可設定辯論輪數（預設 3 輪 + 裁決）

---

## 17. 附錄

### 附錄 A：Go 依賴清單（暫定）

| Package | 用途 | 版本 |
|---------|------|------|
| `github.com/charmbracelet/bubbletea` | TUI 框架 | latest |
| `github.com/charmbracelet/lipgloss` | TUI 樣式 | latest |
| `github.com/charmbracelet/bubbles` | TUI 元件 | latest |
| `github.com/spf13/cobra` | CLI 框架 | latest |
| `github.com/pelletier/go-toml/v2` | TOML 解析 | latest |
| `gopkg.in/yaml.v3` | YAML 解析 | v3 |
| `github.com/stretchr/testify` | 測試斷言 | latest |
| `compress/flate` (stdlib) | PDF FlateDecode | — |
| `encoding/json` (stdlib) | JSON 處理 | — |
| `net/http` (stdlib) | HTTP Client | — |
| `crypto/sha256` (stdlib) | 檔案 hash | — |

### 附錄 B：Agent 角色 YAML 欄位完整定義

| 欄位 | 型別 | 必要 | 預設值 | 說明 |
|------|------|------|--------|------|
| `name` | string | ✅ | — | 唯一識別名（英文、kebab-case） |
| `display_name` | string | ✅ | — | TUI 顯示名稱 |
| `model` | string | ✅ | — | OpenRouter model ID |
| `stance` | enum | ❌ | `neutral` | `critique` / `support` / `neutral` |
| `temperature` | float | ❌ | `0.7` | 0.0 ~ 2.0 |
| `max_tokens` | int | ❌ | 依模型 | 單次回應 token 上限 |
| `focus_areas` | []string | ❌ | `[]` | 關注面向（供 TUI 顯示） |
| `system_prompt` | string | ✅ | — | 系統 prompt |
| `response_template` | string | ❌ | — | 回應格式模板（注入 prompt 尾端） |
| `color` | string | ❌ | 自動分配 | TUI 顏色（hex） |
| `fallback_model` | string | ❌ | — | 主模型不可用時的備用 |

### 附錄 C：指令模式指令完整表

| 指令 | 參數 | 說明 |
|------|------|------|
| `/start` | — | 啟動流水線 |
| `/next` | — | 進入下一輪 |
| `/end` | — | 結束審議 |
| `/pause` | — | 暫停審議 |
| `/focus` | `<start>-<end>` | 聚焦特定行範圍 |
| `/add-file` | `<path>` | 載入補充檔案 |
| `/skip` | `<agent_name>` | 下一輪跳過某 Agent |
| `/summarize` | — | 觸發 Summarizer Agent |
| `/export` | `md` / `pdf` | 匯出目前 session |
| `/agents` | — | 顯示當前 Agent 清單 |
| `/status` | — | 顯示詳細狀態 |
| `/cost` | — | 顯示費用明細 |
| `/help` | — | 顯示指令說明 |
| `/quit` | — | 儲存並退出 |

### 附錄 D：錯誤碼完整表

| 錯誤碼 | 說明 | 使用者可見訊息 |
|--------|------|----------------|
| `CFG-001` | API key 未設定 | 請先執行 `woolf init` 或設定 `OPENROUTER_API_KEY` |
| `CFG-002` | 設定檔格式錯誤 | 設定檔解析失敗，請檢查 {path} |
| `CFG-003` | Agent YAML 格式錯誤 | 角色定義 {name} 解析失敗 |
| `NET-001` | 連線逾時 | 連線 OpenRouter 逾時，請檢查網路 |
| `NET-002` | DNS 解析失敗 | 無法解析 openrouter.ai |
| `API-001` | 401 Unauthorized | API key 無效或已過期 |
| `API-002` | 402 Payment Required | OpenRouter 餘額不足 |
| `API-003` | 429 Rate Limited | 請求過於頻繁，{retry_after} 秒後重試 |
| `API-004` | 模型不可用 | 模型 {model_id} 目前不可用 |
| `API-005` | 500+ Server Error | OpenRouter 伺服器錯誤，正在重試... |
| `ING-001` | 檔案不存在 | 找不到檔案 {path} |
| `ING-002` | 檔案格式不支援 | 不支援 .{ext} 格式，支援的格式：.md, .txt, .pdf |
| `ING-003` | PDF 解析失敗 | PDF 解析錯誤：{detail} |
| `ING-004` | PDF 為加密檔案 | 不支援加密 PDF |
| `ING-005` | PDF 為掃描件 | 不支援掃描件 PDF，請提供含文字層的 PDF |
| `ING-006` | 檔案過大 | 檔案超過 token 上限（{estimated} / {limit}） |
| `SES-001` | Session 不存在 | 找不到 session {id} |
| `SES-002` | 儲存失敗 | Session 儲存失敗：{detail} |
| `SES-003` | JSON 損毀 | Session 檔案已損毀：{path} |
| `TUI-001` | 終端機太小 | 終端機尺寸過小（需 ≥ 80×24） |

---

*End of Spec v1.0*
