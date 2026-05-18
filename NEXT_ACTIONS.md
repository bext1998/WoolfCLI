# WoolfCLI Next Actions

更新日期：2026-05-16

本文件只列 Phase 1 內、可由目前 repo 證據支持的下一步；不擴大到 Phase 2、雲端同步、Web UI、非 OpenRouter provider 或長期 agent memory。

## 最優先的 1 個任務

1. 補齊 ingestion 錯誤碼一致化與測試（Reasoning: Medium）
   - 理由：`.md` / `.txt` / `.pdf` ingestion 是 `woolf start --draft` 的前置路徑，也是 Phase 1 Must Have；目前 `.md` / `.txt` 的錯誤包裝尚未與 `ING-*` 完全一致，且這個任務範圍清楚、跨模組風險低。
   - 建議範圍：`internal/ingestion` 與對應測試；不要同時重寫 PDF parser。
   - 驗收：新增或更新 ingestion 測試，確認不存在檔案、不支援副檔名、`.md` / `.txt` 讀取失敗皆回傳可辨識 `ING-*` 錯誤。

## 接下來可以做的 3～5 個任務

1. 建立 CLI 非互動 smoke flow 測試（Reasoning: Medium）
   - 理由：`start` 已可用 fake client 跑 orchestrator，先把非互動流程鎖住，可降低後續接 TUI 前的回歸風險。
   - 建議範圍：`internal/cli/start_test.go` 與必要 fixture；避免呼叫真實 OpenRouter。

2. 補強 orchestrator 單一 Agent 失敗與 fallback model 策略（Reasoning: High）
   - 理由：規格要求單一 Agent 失敗時 graceful degradation；目前 stream/client error 會標記 skipped/error，但 fallback model 尚未接入，行為也需要更精準測試。
   - 建議範圍：`internal/orchestrator` 測試先行，再評估是否需要擴充 `ChatClient` 或 role resolution。

3. 建立 PDF regression 最小測試集（Reasoning: High）
   - 理由：PDF parser 是高風險區，沒有 fixture 前無法安全改善準確率或錯誤分類。
   - 建議範圍：新增最小 `testdata` / fixture 與 expected Markdown；先涵蓋簡單文字層與明確不支援情境，不處理 OCR。

4. 實作 cost tracker 的最小 session usage 彙總（Reasoning: Medium）
   - 理由：目前 orchestrator 已累計 token usage，但 `internal/cost` 仍是骨架；先把 usage tracker 做穩，再接 pricing / budget warning。
   - 建議範圍：`internal/cost` 與 session/export 相關測試；暫不承諾 OpenRouter dashboard 誤差 < 5%。

5. 設計 TUI 主 model 與 orchestrator event 邊界（Reasoning: High）
   - 理由：TUI 是 Phase 1 核心缺口，但直接開做容易把 UI、session、pipeline 責任混在一起；先定義 Bubble Tea model 如何消費 orchestrator events。
   - 建議範圍：`internal/tui` 設計與小步 smoke test；先做三區佈局和事件顯示，不同時實作所有 `/` 指令。

## 暫時不要做的事

- 不要做 Web UI、雲端同步、外部帳號系統、非 OpenRouter provider 或長期 agent memory。
  - 理由：這些不屬目前 Phase 1 核心，會讓產品方向偏離 `docs/woolf-spec.md` 與 `AGENTS.md`。
- 不要把 PDF export 當成已完成或優先 Must。
  - 理由：目前 CLI 明確只支援 `--format md`，PDF export 在現況中仍應視為 Should / 後續項。
- 不要在沒有 regression fixtures 前大改 PDF parser。
  - 理由：PDF 解析風險高，缺少測試會讓準確率與跨平台行為難以驗證。
- 不要一次重構 CLI、TUI、orchestrator、session schema。
  - 理由：這些都是核心邊界，應以小步任務保留可審查性與相容性。
- 不要新增大型依賴或套件管理方式。
  - 理由：目前 Go module 已有清楚依賴；新增大型依賴需有明確需求與替代方案比較。

## 需要人工確認的事

- 是否需要建立 `PROJECT_BRIEF.md`？（Reasoning: Medium）
  - 理由：本次任務要求讀取該檔，但 repo 目前不存在；若它是專案導航的固定輸入，需確認內容來源與責任邊界。
- Session status 是否維持 `active`，或需要對齊規格中可能出現的 `draft` / `running` 命名？（Reasoning: High）
  - 理由：session JSON 是核心契約，任何 schema / enum 變更都可能影響相容性與既有 session。
- TUI 優先順序是先做 CLI `start` 接 TUI，還是先完成 TUI session browser / layout？（Reasoning: High）
  - 理由：兩者都屬 Phase 1 缺口，但會牽涉不同入口與測試策略。
- PDF Phase 1 驗收樣本應由誰提供？（Reasoning: Medium）
  - 理由：規格要求文字層 PDF 準確率驗收；若沒有代表性中文 / 混合編碼樣本，測試可能失真。
- OpenRouter pricing cache 的資料來源與更新策略是否要先定義？（Reasoning: High）
  - 理由：成本精準度目標會依賴模型定價資料的來源、快取與更新時機。
