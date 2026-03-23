# KubeCaps — AI-Powered Kubernetes Resource & Autoscaling Advisor CLI

A Go CLI tool that predicts optimal resource configurations, evaluates existing HPA/VPA/KEDA setups, and validates them against real Prometheus metrics.

## User Review Required

> [!IMPORTANT]
> **Prometheus Dependency**: The tool assumes a running Prometheus instance accessible from the CLI. Do you have a Prometheus endpoint available, or should we make it optional and support mock data for development?

> [!IMPORTANT]
> **ML Approach**: For resource prediction, we'll use statistical methods (linear regression, EWMA, percentile analysis) implemented in pure Go — no Python/TensorFlow dependency. This keeps the CLI self-contained. Is this acceptable, or do you want a heavier ML integration?

> [!IMPORTANT]
> **Scope**: This is a very large project. I recommend building it incrementally — Phase 1–3 first for a working MVP, then Phase 4–6 for the evaluation engine. Shall I proceed with the full build?

---

## Project Structure

```
KubeCaps/
├── cmd/
│   └── kubecaps/
│       └── main.go                  # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                  # Root cobra command
│   │   ├── analyze.go               # `analyze` command — resource prediction
│   │   ├── evaluate.go              # `evaluate` command — autoscaling evaluation
│   │   └── flags.go                 # Shared CLI flags
│   ├── k8s/
│   │   ├── client.go                # Kubernetes client wrapper
│   │   ├── hpa.go                   # HPA detection & config extraction
│   │   ├── vpa.go                   # VPA detection & config extraction
│   │   └── keda.go                  # KEDA detection & config extraction
│   ├── prometheus/
│   │   ├── client.go                # Prometheus API client
│   │   └── queries.go               # PromQL query templates
│   ├── predictor/
│   │   ├── engine.go                # ML prediction engine
│   │   ├── linear.go                # Linear regression
│   │   ├── ewma.go                  # EWMA (Exponential Weighted Moving Average)
│   │   └── percentile.go            # Percentile-based analysis
│   ├── evaluator/
│   │   ├── hpa.go                   # HPA evaluation logic
│   │   ├── vpa.go                   # VPA evaluation logic
│   │   ├── keda.go                  # KEDA evaluation logic
│   │   ├── correlation.go           # Cross-system correlation analysis
│   │   ├── scorer.go                # Scoring system
│   │   └── risk.go                  # Risk detection engine
│   └── output/
│       ├── formatter.go             # Rich terminal output
│       ├── table.go                 # Table rendering
│       └── colors.go                # Color-coded output helpers
├── pkg/
│   └── types/
│       └── types.go                 # Shared types and models
├── go.mod
├── go.sum
└── README.md
```

---

## Proposed Changes

### Phase 1: Project Scaffolding & Core Infrastructure

#### [NEW] [go.mod](file:///Users/vasudevchavan/Mac-local/KubeCaps/go.mod)
Initialize Go module `github.com/vasudevchavan/kubecaps` with dependencies:
- `github.com/spf13/cobra` — CLI framework
- `k8s.io/client-go` — Kubernetes API client
- `github.com/prometheus/client_golang` — Prometheus HTTP API
- `github.com/fatih/color` — Terminal colors
- `github.com/olekukonez/tablewriter` — Table output

#### [NEW] [types.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/pkg/types/types.go)
Core shared types:
- `WorkloadInfo` — namespace, name, kind, labels
- `ResourceUsage` — CPU/Mem time-series data
- `HPAConfig`, `VPAConfig`, `KEDAConfig` — autoscaling configs
- `Recommendation` — resource/scaling recommendation with confidence
- `EvaluationResult` — evaluation output with score, insights, risks
- `CorrelationResult` — cross-metric correlation data

#### [NEW] [main.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/cmd/kubecaps/main.go)
CLI entry point, calls `cli.Execute()`.

#### [NEW] [root.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/cli/root.go)
Root cobra command with global flags: `--kubeconfig`, `--prometheus-url`, `--namespace`, `--output-format`.

#### [NEW] [flags.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/cli/flags.go)
Shared flag definitions and validation helpers.

---

#### [NEW] [client.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/k8s/client.go)
Kubernetes client wrapper:
- Auto-detects kubeconfig (flag > env > default path)
- Provides typed accessors for Deployments, StatefulSets, DaemonSets
- Lists workloads by namespace

#### [NEW] [client.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/prometheus/client.go)
Prometheus API client:
- `QueryRange()` — range queries for time-series data
- `Query()` — instant queries
- Error handling and retry logic

#### [NEW] [queries.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/prometheus/queries.go)
PromQL query templates for:
- CPU usage by pod/container
- Memory usage (RSS, working set)
- Request rates (RPS)
- HPA scaling events
- OOM events
- Kafka consumer lag (for KEDA)

---

### Phase 2: Resource Prediction Engine

#### [NEW] [engine.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/predictor/engine.go)
Main prediction engine that:
- Fetches historical CPU/Memory from Prometheus
- Runs multiple prediction models
- Aggregates results with weighted confidence
- Returns `Recommendation` with confidence level

#### [NEW] [linear.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/predictor/linear.go)
Linear regression for trend-based CPU/Memory prediction.

#### [NEW] [ewma.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/predictor/ewma.go)
Exponentially weighted moving average for smoothed predictions.

#### [NEW] [percentile.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/predictor/percentile.go)
Percentile-based analysis (P50, P95, P99) for safe request/limit recommendations.

#### [NEW] [analyze.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/cli/analyze.go)
`kubecaps analyze` command:
- Takes workload name/namespace
- Runs prediction engine
- Outputs resource recommendations with scores

---

### Phase 3: Autoscaling Detection & Config Extraction

#### [NEW] [hpa.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/k8s/hpa.go)
HPA detection:
- Lists HPAs targeting the workload
- Extracts: min/maxReplicas, metrics (CPU%, Mem, custom), behavior config

#### [NEW] [vpa.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/k8s/vpa.go)
VPA detection:
- Lists VPAs targeting the workload
- Extracts: update mode, resource policies, bounds

#### [NEW] [keda.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/k8s/keda.go)
KEDA detection:
- Lists ScaledObjects/ScaledJobs targeting the workload
- Extracts: trigger types, thresholds, polling intervals, cooldown

---

### Phase 4: Autoscaling Evaluation Engine

#### [NEW] [hpa.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/evaluator/hpa.go)
HPA evaluation:
- **Scaling Responsiveness** — correlate scaling events with traffic spikes, detect latency
- **Target Accuracy** — compare HPA CPU target vs actual utilization patterns
- **Replica Sufficiency** — detect under-provisioning during peak load
- **Oscillation Detection** — detect frequent scale up/down events
- Returns score (1–10) + insights + risks

#### [NEW] [vpa.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/evaluator/vpa.go)
VPA evaluation:
- **Request vs Actual Gap** — compare VPA recommendations to actual usage
- **OOM Risk** — analyze memory limits vs peak usage patterns
- **Throttling Detection** — check CPU throttling metrics
- Returns score (1–10) + insights + risks

#### [NEW] [keda.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/evaluator/keda.go)
KEDA evaluation:
- **Trigger Accuracy** — compare trigger activation vs actual load arrival
- **Threshold Correctness** — validate thresholds against metric patterns
- **Scaling Lag** — measure delay between trigger and pod ready
- Returns score (1–10) + insights + risks

---

### Phase 5: Cross-System Intelligence

#### [NEW] [correlation.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/evaluator/correlation.go)
Correlation analysis:
- Cross-correlate traffic (RPS), CPU/Memory, and scaling events
- Identify causal relationships (e.g., "CPU driven by traffic, not memory")
- Generate actionable insights

#### [NEW] [scorer.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/evaluator/scorer.go)
Scoring system:
- Per-component scores (HPA, VPA, KEDA: 1–10)
- Overall "Optimization Score" combining all evaluations
- Weight factors for different risk categories

#### [NEW] [risk.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/evaluator/risk.go)
Risk detection:
- OOM risk flagging
- CPU throttling detection
- Scaling lag warnings
- Over-provisioning alerts

---

### Phase 6: CLI UX & Output

#### [NEW] [evaluate.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/cli/evaluate.go)
`kubecaps evaluate` command with flags:
- `--evaluate-autoscaling` — run full evaluation
- `--compare-current` — show current vs recommended config table
- `--show-metrics` — include raw Prometheus metric evidence
- `--explain` — verbose explainability output

#### [NEW] [formatter.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/output/formatter.go)
Rich output formatting:
- Color-coded scores (green/yellow/red)
- Insight blocks with emoji indicators
- Confidence level visualization

#### [NEW] [table.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/output/table.go)
Comparison table renderer for current-vs-recommended configs.

#### [NEW] [colors.go](file:///Users/vasudevchavan/Mac-local/KubeCaps/internal/output/colors.go)
Terminal color helpers and theme definitions.

---

## Verification Plan

### Automated Tests

1. **Unit tests** for prediction models:
   ```bash
   go test ./internal/predictor/... -v
   ```
   - Test linear regression, EWMA, percentile on known datasets
   - Verify confidence scoring

2. **Unit tests** for evaluators:
   ```bash
   go test ./internal/evaluator/... -v
   ```
   - Test HPA/VPA/KEDA evaluation logic with mock Prometheus data
   - Test scoring system calculations
   - Test risk detection thresholds

3. **Full test suite**:
   ```bash
   go test ./... -v
   ```

4. **Build verification**:
   ```bash
   go build -o kubecaps ./cmd/kubecaps
   ```

### Manual Verification

1. **CLI Help** — Run `./kubecaps --help`, `./kubecaps analyze --help`, `./kubecaps evaluate --help` and verify command structure and flags
2. **Against a live cluster** (if available) — Run `./kubecaps analyze --namespace default` to verify Kubernetes client and Prometheus integration work end-to-end
3. **Output formatting** — Visually verify tables, colors, and score formatting look correct in terminal
