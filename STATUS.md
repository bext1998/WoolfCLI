# WoolfCLI Status

更新日期：2026-05-16

本文件依據 `README.md`、`docs/woolf-spec.md`、`docs/testing.md`、`docs/todo.md`、目前程式碼結構與工作區狀態整理。`PROJECT_BRIEF.md` 在本次檢查時不存在；以下不引用該檔內容。

## 目前完成了什麼

- Go module、Cobra CLI、主要 internal 模組與測試腳本已建立。
- CLI 根命令目前接線 `init`、`start`、`resume`、`list`、`show`、`export`、`fork`、`delete`、`agents`、`config`、`models`。
- `woolf start` 目前會載入 config/runtime dirs、套用 preset 或指定 agents、必要時 ingestion `--draft`、建立 session、執行 orchestrator pipeline，並輸出 agent started / finished 與 session done 事件。
- Agent 系統已有 6 個內建角色與 `editorial`、`brainstorm`、`critique`、`review` preset；`agents` 指令已涵蓋 list/show/add/delete/validate 與 preset 子命令。
- OpenRouter client 已有 Chat Completions SSE 串流解析、usage 解析、API key 缺失檢查、HTTP 401/402/429/404/5xx 錯誤碼映射，以及 429 / 5xx bounded retry 測試。
- Orchestrator 已能依序執行多 Agent，後序 Agent context 會包含前序回覆，並把 response、token usage、session totals 持久化。
- Context builder 已將 draft、summaries、interventions、focus range、role prompt metadata、response template 與 previous discussion 納入 messages。
- Session store 已支援 create/save/load/resume/fork/delete/list，並有 JSON parse error、session persistence、cancellation、stream error 等測試覆蓋。
- Ingestion 已有 `.md`、`.txt` 與 `.pdf` 入口；PDF 目前屬 Phase 1 基礎解析範圍。
- Markdown export 已有入口；CLI `export` 目前只支援 `--format md`。
- Config 已支援 TOML defaults、runtime path、`OPENROUTER_API_KEY` / `WOOLF_SESSIONS_DIR` env override 與 API key masking。
- CI 設定目前在 Ubuntu / Windows 上執行 `go mod download`、`go test ./...`、`go vet ./...`。

## 目前缺什麼

- TUI 仍主要是骨架，尚未完成規格要求的三區佈局、串流渲染、指令模式、長文輸入驗證、小視窗 fallback 與主流程接線。
- `woolf start` 還不是完整互動式 TUI 主流程，目前偏非互動 CLI pipeline 執行。
- Orchestrator 尚缺明確 state machine、interactive intervention loop、`/skip` 接入、fallback model 與完整 graceful degradation 行為。
- Context 尚缺 token 預估、context window 裁切、summary compression，以及 `/focus` 對稿件行範圍的精準切片。
- Ingestion 尚缺 `.md` / `.txt` 錯誤碼一致化、PDF regression fixtures、PDF 文字層準確率驗收，以及掃描 / 加密 / 無文字層 PDF 的明確 graceful error。
- Cost 目前只有骨架，尚缺 per-agent / per-round / per-session cost tracker、OpenRouter pricing table、budget warning 與 dashboard 誤差驗收。
- Markdown export 尚需完整驗收 metadata、agents、rounds、stance、interventions、tokens、cost 格式；PDF export 仍不是 Phase 1 已落地功能。
- Error handling 尚缺 `ING-*`、`NET-*`、`TUI-*` 的一致錯誤型別與測試，以及 log 系統與敏感資訊遮蔽整體驗證。
- `models` 指令目前只輸出 models cache 路徑；`--pricing` 仍是占位訊息，不是已完成 pricing 功能。

## 已知 bug

- 未確認：目前沒有 issue tracker 或 bug report 檔案可供交叉比對。
- 已知實作缺口：`docs/todo.md` 標示多個 Phase 1 未完成項，尤其 TUI、PDF regression、cost、context compression 與 graceful degradation。
- 相容性疑慮：session status 目前使用 `active`、`paused`、`completed`、`error`；`docs/todo.md` 記錄需評估是否與規格中的狀態命名完全對齊。
- 產品行為疑慮：stance tag 目前由 role stance 推導，尚未從 Agent 回覆解析或驗證。
- 測試覆蓋疑慮：PDF、TUI、自動化 E2E、cost 精準度與部分 CLI 行為仍缺明確驗收。

## 最近變動

- 工作區在本次檢查前已有未提交變更：`AGENTS.md`、`internal/cli/agents.go`、`internal/cli/agents_test.go`。
- `AGENTS.md` 的既有變更補充了目前 CLI 指令面、agents/config/models/start/export 的實際限制，以及測試工作流說明。
- `internal/cli/agents.go` 的既有變更讓 `agents add` 寫入原始 YAML 內容，而不是重新 marshal role struct。
- `internal/cli/agents_test.go` 的既有變更加入測試，確認自訂 YAML 中額外欄位可被保留。
- 本次 Project Navigation 新增 `STATUS.md` 與 `NEXT_ACTIONS.md`，不修改程式碼。

## 風險或疑慮

- 規格範圍大於目前實作，尤其 TUI、成本估算、PDF Phase 1 與 graceful degradation；若下一步未聚焦，容易擴散成多模組大改。
- `docs/woolf-spec.md` 第 11 章描述的 CI pipeline 包含 lint、race、cover、build、PDF regression，但目前 repo 的 GitHub Actions 只跑 download/test/vet；文件或任務規劃需避免把未存在 CI job 當成事實。
- PDF 自建解析器仍是高工程風險區；沒有 regression fixtures 前不適合大量修改 parser。
- Cost 功能會影響 TUI、session totals、export 與 OpenRouter pricing cache；在 usage tracker 未穩定前，不宜承諾 dashboard 誤差目標。
- TUI 是 Phase 1 核心但目前落差最大；若直接接 OpenRouter 串流與鍵盤命令，需先界定 Bubble Tea model、event 與 orchestrator 邊界。
- 未確認：`PROJECT_BRIEF.md` 缺失是刻意未建立，還是應補上的專案文件。
