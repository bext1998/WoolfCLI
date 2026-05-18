# WoolfCLI 下一步

更新日期：2026-05-17

本文件只列下一次開工最合理的任務順序。依據目前 repo 狀態，下一步應優先降低規格不確定性，然後補 Phase 1 最大缺口。無法確認的內容標記為「未確認」。

## 最合理的下一步

1. 先補齊 TUI Phase 1 最小可驗收主流程。
   - 範圍：`internal/tui`，以及必要的 `internal/orchestrator` 介面串接。
   - 目標：讓使用者能在 TUI 中看到討論、輸入介入內容、觸發基本命令、看見狀態與串流輸出。
   - 規格依據：`docs/spec.md` M-05、M-06、M-07、M-16、M-17、7.11、14.1。
   - 驗收：基本 keyboard workflow、discussion/input/status 區塊、streaming delta、`/help` 或基本 slash command parser、視窗過小 fallback。
   - 注意：TUI 不應直接寫 session 檔案，session persistence 仍應經 orchestrator/session 邊界處理。

## 接續任務

2. 補齊 orchestrator intervention loop 與 graceful degradation。
   - 範圍：`internal/orchestrator`。
   - 目標：補 state machine、輪次間介入、`/skip`、fallback model、單一 agent 失敗後繼續、多 agent 失敗後中止與可診斷 session 狀態。
   - 規格依據：M-01、M-02、M-16、7.4、9.3、14.1。

3. 補 ingestion 錯誤路徑與 PDF regression。
   - 範圍：`internal/ingestion`、`internal/ingestion/pdf`、新增最小必要 `testdata/pdf` fixture。
   - 目標：覆蓋 `.md`、`.txt`、具文字層 PDF 成功路徑，以及不支援 PDF 的 `ING-*` graceful error。
   - 注意：不得宣稱支援 OCR、掃描 PDF、加密 PDF 或複雜版面還原。
   - 未確認：目前 PDF parser 對真實 fixture 的解析品質。

4. 補 Markdown export 測試與格式驗收。
   - 範圍：`internal/exporter` 與必要 CLI 測試。
   - 目標：確認 Markdown export 包含 metadata、source、agents、rounds、stance、interventions、tokens、cost，且格式符合規格第 7.8 節。
   - 注意：PDF export 目前不應搶先於 Markdown export 驗收，除非人工重新排定優先級。

5. 補 cost tracker 實際流程。
   - 範圍：`internal/cost`、orchestrator token usage、CLI/TUI 顯示。
   - 目標：per-agent/per-round/per-session token 與 cost summary，並處理 pricing cache 或「定價可能不準確」提示。
   - 未確認：是否已有可用 OpenRouter pricing fixture 或應先用硬編碼測試資料。

6. 同步 README 狀態。
   - 範圍：`README.md`。
   - 目標：把「早期規劃與實作前階段」更新為符合目前 repo 的狀態。
   - 注意：本次任務限制只更新三份文件，因此 README 先列為後續。

7. 強化測試管線。
   - 範圍：`scripts/test.ps1`、`.github/workflows/go-test.yml`、必要 fixture。
   - 目標：在既有 `go test ./...` / `go vet ./...` 之外，視需要補 race、coverage、PDF regression 的可重現流程。
   - 未確認：race/coverage 是否要成為每次 CI 必跑，或只作為手動/夜間檢查。

## 需要人工決定的事項

- 是否正式接受 `API-001`、`API-002`、`API-003`、`API-004`、`API-005` 作為 OpenRouter/API error code 對外契約。
- 下一個工程焦點要先完成 TUI 可驗收主流程，或先完成 CLI-only pipeline 的 export/cost/ingestion 驗收。
- PDF export 是否仍維持 Should Have / Phase 2，還是要提前納入近期開發。
- 是否允許在下一次任務同步更新 README，修正目前「實作前階段」的過期描述。
- 是否有必要建立固定 PDF regression fixture；若 fixture 可能含版權或敏感內容，需要提供可公開測試樣本。

## 暫不建議開始

- 不要新增 Web UI、雲端同步、資料庫後端、多人協作或非 OpenRouter provider。
- 不要先做大規模 TUI 美化；應先完成可驗收互動與測試。
- 不要先做 PDF export，除非已確認它比 PDF ingestion regression 更優先。
- 不要更動 session schema 或 CLI command output，除非同步處理相容性、測試與文件。
- 不要引入大型新依賴；若 TUI、PDF 或 cost 需要依賴，應先提出替代方案與影響範圍。
