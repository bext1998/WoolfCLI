# AGENTS.md

本文件是 WoolfCLI 專案中 Codex 與所有子代理的工作契約。任何開發、修正、重構、測試、文件更新與審查工作，都必須以 `docs/woolf-spec.md` 作為最高規格來源，並遵循本文的流程限制，避免未經證據支持的指示飄移。

## 0. 優先順序

當指令、規格或現有程式碼出現衝突時，依以下順序判斷：

1. 使用者在當前任務中的明確要求。
2. 本檔案 `AGENTS.md`。
3. `docs/woolf-spec.md`。
4. 現有程式碼、測試、README、`docs/testing.md` 與檔案命名慣例。
5. 一般工程最佳實務。

不得用一般最佳實務覆蓋 Woolf 規格。若規格缺漏，必須先從現有程式碼找證據；仍無法判斷時，提出明確假設並控制修改範圍。

## 1. Harness Engineering 工作模型

本專案採用 harness engineering 式 AI 協作流程：先設計代理的工作環境與約束，再讓代理實作。核心原則是「結構化上下文輸入，產生可預測輸出」。

所有非微小變更都必須分成三個階段：

### Phase 1：Repository Impact Map

開始修改前，代理必須先閱讀真實檔案並建立任務影響地圖。影響地圖至少包含：

- 任務目標：本次要完成的可驗收結果。
- 規格依據：引用 `docs/woolf-spec.md` 的相關章節或條目，例如 M-01、7.4、8、9.2、11。
- 既有入口：相關 CLI、TUI、orchestrator、session、agent、openrouter、ingestion、exporter、config 或 cost 模組。
- 預計修改檔案：只能列出真實存在的路徑；若要新增檔案，需說明責任邊界。
- 既有模式：要沿用的函式、型別、介面、錯誤處理或測試寫法。
- 驗收方式：要執行的測試、手動檢查或無法執行時的替代驗證。

禁止在沒有讀取相關檔案的情況下直接實作。禁止憑空發明路徑、型別、API、命令或產品行為。

### Phase 2：Structured Task Contract

進入實作前，代理必須把任務收斂成結構化契約：

- Repository：`D:\AgentCoding\WoolfCLI`
- Description：本次變更的單一目的。
- Files to Modify：精準列出本次會修改或新增的檔案。
- Implementation Notes：引用實際符號、既有測試或既有架構模式。
- Acceptance Criteria：列出可驗收條件，需對應 Woolf 規格或使用者要求。
- Test Requirements：列出必跑測試；若跳過，需說明原因。

子代理只能在被指派的檔案與責任邊界內工作。不得因為看到其他問題而順手重構或改動未指派範圍。

### Phase 3：Implementation and Verification

實作時必須遵守：

- 小步修改，保持差異可審查。
- 每個修改都要能回扣到任務契約或規格條目。
- 優先修正根因，不用註解掉錯誤、吞掉錯誤或跳過流程來假裝完成。
- 實作完成後必須執行相稱測試，預設至少執行 `go test ./...` 或 `.\scripts\test.ps1`。
- 若變更涉及 CLI 指令面、orchestrator 流程、session schema、OpenRouter client、ingestion、export 或 cost 計算，必須補上或更新對應測試。

## 2. Woolf 產品不可偏移原則

Woolf 是「面向文字創作者與內容工作者的多模型 AI 審議 CLI/TUI」，不是通用聊天機器人、IDE agent、專案管理工具、RAG 平台或網頁應用。

所有功能必須維持以下核心：

- 使用 Go 實作 CLI/TUI。
- 透過 OpenRouter API 調度多個 AI Agent。
- 多 Agent 依序發言，後發言者可讀取前序完整內容。
- 支援 `agree`、`disagree`、`extend`、`neutral` 立場標籤。
- Session 是持久化單位，必須可儲存、查詢、續接與匯出。
- 檔案 ingestion 以 `.md`、`.txt`、具文字層 `.pdf` 為 Phase 1 範圍。
- TUI 必須服務長文創作工作流，保持鍵盤友善、低摩擦、可中斷與可續接。

不得新增與規格無關的產品方向，例如雲端同步、多人協作、Web UI、資料庫後端、外部帳號系統、非 OpenRouter provider、長期 agent memory 或 Phase 2 功能，除非使用者明確要求。

## 3. 架構邊界

必須維持現有模組責任：

- `cmd/woolf`：CLI 程式入口，只做啟動與根命令接線。
- `internal/cli`：Cobra 指令、參數解析、命令流程與使用者可見錯誤。
- `internal/tui`：Bubble Tea TUI、視圖、元件、鍵盤操作與畫面狀態。
- `internal/orchestrator`：Agent 流水線、狀態機、context 組裝、介入點、取消與壓縮策略。
- `internal/agents`：角色、prompt、preset、YAML 載入與 registry。
- `internal/openrouter`：OpenRouter API、SSE 串流、模型清單、rate limit 與 API 錯誤映射。
- `internal/ingestion`：稿件讀取與格式轉換；PDF 子模組限於 PDF 解析與 Markdown 化。
- `internal/session`：Session schema、store、resume、fork、search 與持久化。
- `internal/exporter`：Markdown/PDF 匯出，不負責 session 決策。
- `internal/cost`：token 與費用估算。
- `internal/config`：TOML 設定、路徑與環境變數載入。
- `pkg/pdfparse`：可公開復用的 PDF parsing 型別或工具。

禁止跨層塞責任。例如：

- 不可在 CLI 指令中直接組 OpenRouter request。
- 不可在 TUI 元件中寫 session 檔案。
- 不可在 orchestrator 中硬編 agent preset。
- 不可在 openrouter client 中處理 TUI 顯示邏輯。
- 不可在 ingestion 中做 agent prompt 或 session 決策。

## 4. 規格驅動開發守則

實作前必須核對 `docs/woolf-spec.md`，尤其是：

- 第 5 章功能規格：Must/Should/Could 範圍。
- 第 6 章系統架構：模組關係與資料流。
- 第 7 章技術規格：CLI、OpenRouter、Agent、Orchestrator、Context、Ingestion、Session、Export、Cost、Config、TUI。
- 第 8 章資料格式定義：session JSON 與 agent YAML schema。
- 第 9 章錯誤處理策略：錯誤分類與 graceful degradation。
- 第 10 章安全性：API key、session、HTTPS、日誌與依賴。
- 第 11 章測試策略：單元、整合、端到端與 PDF regression。
- 第 14 章 KPI 與驗收標準。

若現有程式碼與規格不一致：

- 任務是「補齊規格」時，以規格為準。
- 任務是「修 bug」時，先確認目前測試與行為是否已形成事實契約，再決定最小修正。
- 涉及 session schema、CLI 命令輸出、錯誤碼或使用者工作流時，必須在回報中明確指出相容性影響。

## 5. Go 開發規範

- 使用 Go 1.22 或更新版本。
- 優先沿用標準函式庫與既有依賴，不得未經需求引入大型依賴。
- 新增依賴前必須說明原因、替代方案與影響範圍。
- 對外行為使用明確錯誤回傳，不用 panic 處理可預期錯誤。
- 錯誤訊息應可對應規格中的錯誤類別，例如 `CFG-*`、`NET-*`、`API-*`、`ING-*`、`SES-*`、`TUI-*`。
- 優先使用 interface 隔離外部 API、檔案系統與時間等副作用，方便測試。
- 不做大規模重排、格式化無關檔案或跨模組命名風格切換。

## 6. OpenRouter 與安全要求

- 不可硬編 API key、token、密碼或測試用真實憑證。
- API key 僅能從環境變數或本機設定讀取。
- CLI、TUI、log、測試 fixture 中顯示 key 時必須遮蔽。
- OpenRouter API 錯誤必須映射為規格定義的可理解錯誤與 graceful degradation 行為。
- 429 必須尊重 retry-after 或既有 retry 策略。
- 500-503 類錯誤應使用有限次數退避重試，不得無限重試。
- 測試不得呼叫真實 OpenRouter，必須使用 fake client、mock transport 或 fixture。

## 7. Session 與資料格式要求

- Session JSON 是產品核心契約；任何 schema 變更都必須謹慎。
- 不得刪除既有 session 欄位，除非任務明確要求並提供 migration 或相容策略。
- 每輪完成後應可自動保存；中斷、取消、stream error 也必須留下可診斷狀態。
- `stance_tag` 僅能使用 `agree`、`disagree`、`extend`、`neutral` 或規格允許的 null。
- Session store 測試應覆蓋儲存、讀取、續接、錯誤檔案與權限情境。
- 匯出功能只能讀取 session 結果，不應改寫 session 決策或內容。

## 8. CLI/TUI 行為要求

- CLI 命令必須符合規格附錄與 README 已公開的命令面。
- 目前 root CLI 以 `internal/cli/root.go` 為準，已實作 `init`、`start`、`resume`、`list`、`show`、`export`、`fork`、`delete`、`agents`、`config`、`models`；更新 `AGENTS.md`、README 或測試時不得憑空加入不存在的指令。
- `agents` 目前已實作 `list`、`show <name>`、`add <role-yaml>`、`delete <name>`、`validate [role-yaml...]`，以及 `agents preset list|show <name>`；涉及 agent 管理流程時應以這組命令面為依據。
- `config` 目前已實作 `show`、`reset`、`edit`；`models` 目前僅輸出 models cache 路徑，`--pricing` 只顯示占位訊息，不能在文件或流程說明中描述成已完成的 pricing 功能。
- 需要新增或修改命令時，必須同步更新 help、測試與必要文件。
- `start` 的實際流程是：載入 config 與 runtime 目錄、套用預設 preset / rounds、必要時 ingestion `--draft`、建立 session、執行 orchestrator pipeline，並輸出 agent started / finished 與 session done 事件；相關修改應沿用這條流程與輸出節點。
- `export` 目前 Phase 1 只支援 `--format md`；若工作涉及匯出行為，不可把 PDF 匯出當成已落地功能。TODO：若未來實作 PDF export，需同步更新本檔、README 與測試。
- TUI 必須鍵盤友善，支援長文輸入與串流輸出，不可讓 5000 字輸入明顯卡頓。
- TUI 畫面不應暴露敏感資訊。
- `/start`、`/next`、`/end`、`/pause`、`/focus`、`/add-file`、`/skip`、`/summarize`、`/export`、`/agents`、`/status`、`/cost`、`/help`、`/quit` 等行為需依規格保持一致。
- 對不支援的終端或過小視窗，應提供可理解 fallback 或錯誤訊息。

## 9. Ingestion 與 PDF 要求

- `.md` 與 `.txt` 應保持低風險直讀，尊重 UTF-8 與既有檔案錯誤處理。
- PDF Phase 1 僅承諾具文字層 PDF 的基本解析與 Markdown 輸出。
- 不得假裝支援掃描 PDF、OCR、加密 PDF 或複雜版面還原，除非規格或任務明確要求。
- PDF 解析失敗時，應回傳 `ING-*` 類錯誤並保留可診斷 detail。
- PDF 測試優先使用 `testdata/pdf` 與 expected fixture；新增 fixture 要最小化且有明確用途。

## 10. 測試與驗收

預設驗證命令：

```powershell
.\scripts\test.ps1
```

可用替代命令：

```powershell
go test ./...
go vet ./...
.\scripts\test.ps1 -Vet
.\scripts\test.ps1 -Race
.\scripts\test.ps1 -Coverage
```

具體工作流補充：

- `.\scripts\test.ps1` 會先執行 `go mod download`，預設再跑 `go test ./...`；`-Coverage` 會改跑 coverage 並輸出 `coverage.out` 摘要，`-Race` 與 `-Vet` 會追加對應檢查。
- 在支援 `make` 的 shell 中，可使用 `make test`、`make test-vet`、`make test-race`、`make test-cover`；更新文件時應與 `Makefile`、`docs/testing.md` 保持一致。
- 現行 CI 依 `docs/testing.md` 與 repo 設定，至少覆蓋 Ubuntu 與 Windows 的 `go mod download`、`go test ./...`、`go vet ./...`；涉及測試流程描述時，不可把未存在的 CI job 當成事實。

測試策略：

- CLI：命令存在性、參數解析、錯誤輸出、無 API key 情境。
- Orchestrator：pipeline 順序、context 傳遞、取消、stream error、session persistence。
- OpenRouter：request 組裝、SSE parsing、rate limit、錯誤映射。
- Agents：built-in role、preset、YAML 載入、schema validation。
- Session：store、resume、fork、search、schema 相容性。
- Ingestion：md/txt/pdf 成功與失敗路徑。
- Exporter：Markdown/PDF 輸出格式與錯誤處理。
- Cost：token 與費用估算誤差控制。
- Smoke coverage 目前已有明確對象：`internal/cli/root_test.go` 驗證 root command surface，`start` 測試使用 fake client 走 orchestrator，config 測試覆蓋 env override 與 API key masking，OpenRouter 測試覆蓋 SSE / 429 / 5xx，orchestrator 測試覆蓋 persistence、cancellation、stream error 與 context propagation。

若無法執行測試，最終回報必須說明原因、未驗證風險與建議補驗方式。

## 11. 子代理協作規則

分派子代理時，必須提供結構化任務契約，且明確指定：

- 可修改檔案或目錄。
- 不可修改的檔案或介面。
- 必須遵守的 `docs/woolf-spec.md` 條目。
- 需要回報的測試結果。

子代理回報必須包含：

- 實際修改檔案。
- 對應規格或任務要求。
- 主要邏輯變更。
- 已執行測試。
- 未解風險或假設。

主代理整合時必須重新檢查子代理變更是否偏離規格、是否跨越責任邊界、是否引入未要求依賴或產品方向。

## 12. 禁止事項

- 不得在未讀規格與相關程式碼前開始實作。
- 不得憑空新增需求、命令、資料欄位、模型 provider 或 UI 流程。
- 不得將 Phase 2 或創意延伸功能當成 Phase 1 必做項。
- 不得為了通過測試而降低測試品質、刪除測試、跳過錯誤或吞掉失敗。
- 不得提交真實 API key、session 私密內容、使用者稿件或敏感 fixture。
- 不得大規模重構無關模組。
- 不得把 unrelated formatting、dependency churn、go.sum churn 混入功能修正。
- 不得破壞 Windows、macOS、Linux/WSL 目標平台相容性。

## 13. 最終回報格式

每次完成任務後，回報必須使用繁體中文，並包含：

- 改了哪些檔案。
- 為什麼這樣改。
- 主要邏輯變更是什麼。
- 執行了哪些測試與結果。
- 可能風險或後續注意事項。

若沒有修改程式碼，也必須清楚說明做了哪些檢查與結論。

