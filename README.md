# 🧢 KubeCaps

**AI-Powered Kubernetes Resource & Autoscaling Advisor & Auditor**

KubeCaps is a CLI tool designed to help Kubernetes operators optimize their workloads by predicting optimal resource configurations and auditing existing autoscaling setups (HPA, VPA, KEDA) against real Prometheus metrics.

---

## 🚀 Key Features

- **🔍 Resource Prediction Engine**: Uses pure Go implementations of Linear Regression, EWMA (Exponential Weighted Moving Average), and Percentile Analysis (P50/P95/P99) to recommend optimal CPU/Memory requests and limits.
- **⚖️ Autoscaling Auditor**: Evaluates HPA, VPA, and KEDA configurations against actual usage patterns to detect scaling lag, oscillation, and target mismatches.
- **🏆 Optimization Scoring**: Provides a 1–10 "Optimization Score" and literal grades (A+ to F) for each component and the overall workload.
- **🚨 Risk Detection**: Automatically flags OOM risks, CPU throttling, over-provisioning, and scaling bottlenecks.
- **📈 Cross-System Intelligence**: Performs Pearson correlation analysis between traffic (RPS), resource usage, and scaling events to identify causal relationships.
- **🎨 Rich Terminal UX**: Beautiful, color-coded reports with score bars, emojis, and Unicode tables.

---

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/vasudevchavan/kubecaps.git
cd kubecaps

# Install dependencies
go mod tidy

# Build the binary
go build -o kubecaps ./cmd/kubecaps

# Move to your path (optional)
mv kubecaps /usr/local/bin/
```

---

## 🛠 Usage

### 1. Resource Analysis
Predict optimal CPU/Memory requests and limits for a specific workload or an entire namespace.

```bash
# Analyze a specific deployment
kubecaps analyze my-app --prometheus-url http://prometheus:9090 -n production

# Analyze all workloads in a namespace
kubecaps analyze --prometheus-url http://prometheus:9090 -n staging
```

### 2. Autoscaling Evaluation
Run a comprehensive audit of HPA, VPA, and KEDA configurations.

```bash
# Full evaluation of a workload
kubecaps evaluate my-app --prometheus-url http://prometheus:9090 -n production

# View JSON output for automation
kubecaps evaluate my-app --prometheus-url http://prometheus:9090 -o json > report.json
```

---

## 📊 Scoring System

KubeCaps uses a weighted scoring mechanism:
- **HPA (35%)**: Evaluates scaling responsiveness, target accuracy, and replica sufficiency.
- **VPA (35%)**: Audits request vs. actual usage gaps, OOM risk, and throttling.
- **KEDA (30%)**: Validates trigger accuracy, thresholds, and polling/cooldown efficiency.

**Grades:**
- `9.0 - 10.0`: **A+** (Optimized)
- `8.0 - 8.9` : **A** (Good)
- `7.0 - 7.9` : **B** (Moderate)
- `Below 6.0`: **F** (Action Required)

---

## 🏗 Architecture

```text
internal/
├── predictor/   # ML Prediction Models (Linear, EWMA, Percentiles)
├── evaluator/   # Audit Engines (HPA, VPA, KEDA, Correlation)
├── k8s/         # Kubernetes Discovery (Typed & Dynamic Clients)
├── prometheus/  # Metrics Integration (15+ Query Templates)
└── output/      # Rich CLI Formatting & UX
```

---

## 🛡 License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.