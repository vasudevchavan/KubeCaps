# 🧢 KubeCaps

[![Go Version](https://img.shields.io/github/go-mod/go-version/vasudevchavan/kubecaps)](https://go.dev/)
[![CI Status](https://github.com/vasudevchavan/kubecaps/workflows/CI/badge.svg)](https://github.com/vasudevchavan/kubecaps/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vasudevchavan/kubecaps)](https://goreportcard.com/report/github.com/vasudevchavan/kubecaps)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![codecov](https://codecov.io/gh/vasudevchavan/kubecaps/branch/main/graph/badge.svg)](https://codecov.io/gh/vasudevchavan/kubecaps)

**AI-Powered Kubernetes Resource & Autoscaling Advisor & Auditor**

KubeCaps is a CLI tool designed to help Kubernetes operators optimize their workloads by predicting optimal resource configurations and auditing existing autoscaling setups (HPA, VPA, KEDA) against real Prometheus metrics.

## 📑 Table of Contents
- [Key Features](#-key-features)
- [Installation](#-installation)
- [Usage](#-usage)
- [Configuration](#-configuration--customization)
- [Scoring System](#-scoring-system)
- [Algorithms & Calculations](#-algorithms--calculations)
- [Architecture](#-architecture)
- [Testing](#-testing)
- [Documentation](#-documentation)
- [Roadmap](#-roadmap)

---

## 🚀 Key Features

- **🔍 Resource Prediction Engine**: Uses pure Go implementations of Linear Regression, EWMA (Exponential Weighted Moving Average), and Percentile Analysis (P50/P95/P99) to recommend optimal CPU/Memory requests and limits.
- **📊 Adaptive Data Analysis**: Automatically identifies and pulls the optimal historical data window (up to 1 Month, adjusting to 1 Week or 1 Day) depending on Prometheus data availability.
- **💡 Contextual Recommendations**: Displays current resource allocations alongside recommendations, computing exact over-provisioned or under-provisioned percentage differences to clarify optimization impact.
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

### 3. AI-Powered Analysis (NEW! 🤖)
Use AI/ML techniques for advanced workload analysis with autoscaler correlation.

```bash
# Full AI analysis with autoscaler correlation
kubecaps ai my-app --prometheus-url http://prometheus:9090 -n production

# Adjust anomaly detection sensitivity
kubecaps ai my-app --prometheus-url http://prometheus:9090 --sensitivity 1.5

# Disable autoscaler analysis (anomalies + DNA only)
kubecaps ai my-app --prometheus-url http://prometheus:9090 --analyze-autoscalers=false

# Get JSON output for automation
kubecaps ai my-app --prometheus-url http://prometheus:9090 -o json
```

**AI Features:**
- 🔍 **Anomaly Detection**: Identifies unusual resource usage patterns using statistical methods (Modified Z-Score, IQR, MAD)
- 🧬 **Workload DNA**: Creates behavioral fingerprints including seasonality, volatility, and growth patterns
- 📊 **Pattern Recognition**: Detects hourly, daily, and weekly patterns automatically using autocorrelation
- 🎯 **Predictability Scoring**: Measures how predictable your workload behavior is (R-squared)
- ⚙️ **Autoscaler Correlation** (NEW!): Analyzes HPA/VPA/KEDA configurations and correlates with actual usage
  - Detects if HPA targets align with actual CPU/Memory usage
  - Compares VPA recommendations vs real resource consumption
  - Validates KEDA polling intervals and trigger thresholds
  - Provides AI-driven insights for autoscaler tuning

**Example Output:**
```bash
$ ./bin/kubecaps ai --prometheus-url=http://localhost:9090 nginx-deployment-b6d4bbc6b-599kn -n default --verbose
🤖 AI Analysis for Deployment/nginx-deployment (default)

🔍 Anomaly Detection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ No CPU anomalies detected

✅ No memory anomalies detected

🧬 Workload DNA Profile
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  Warning: Could not generate DNA profile: insufficient data points: need at least 100, got 8
⚙️  Autoscaler Correlation Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  🧮 KubeCaps Optimization Math (Verbose):

     CPU Strategy:        Baseline(P95: 0.000 cores) → Optimized for risk up to Peak(0.000)*1.4x
     CPU Risk Penalty:    10.0 (Lambda) * 1.5 (SLA) = 15.0 (Total Penalty)
     Mem Strategy:        Baseline(P99: 3.8 MBi) → Optimized for risk up to Peak(3.8 MBi)*1.6x
     Mem Risk Penalty:    30.0 (Lambda) = 30.0 (Total Penalty)
```

**Additional Example:**
```bash
$ ./bin/kubecaps ai --prometheus-url=http://localhost:9090 nginx-deployment-b6d4bbc6b-599kn -n default --verbose
🤖 AI Analysis for Deployment/nginx-deployment (default)

🔍 Anomaly Detection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ No CPU anomalies detected

✅ No memory anomalies detected

🧬 Workload DNA Profile
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  Warning: Could not generate DNA profile: insufficient data points: need at least 100, got 8
⚙️  Autoscaler Correlation Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  🧮 KubeCaps Optimization Math (Verbose):

     CPU Strategy:        Baseline(P95: 0.000 cores) → Optimized for risk up to Peak(0.000)*1.4x
     CPU Risk Penalty:    10.0 (Lambda) * 1.5 (SLA) = 15.0 (Total Penalty)
     Mem Strategy:        Baseline(P99: 3.8 MBi) → Optimized for risk up to Peak(3.8 MBi)*1.6x
     Mem Risk Penalty:    30.0 (Lambda) = 30.0 (Total Penalty)
```

### 4. Configuration & Customization (NEW! ⚙️)
KubeCaps allows you to fine-tune its prediction engine through a YAML configuration file. This is especially useful for adjusting the "Cost vs. Risk" trade-offs for your specific workloads.

```bash
# Run analysis with a custom configuration
kubecaps analyze my-app --config ./kubecaps.yaml
```

Example `kubecaps.yaml`:
```yaml
optimization:
  # Base cost penalty for OOM (Out Of Memory) risk.
  # Higher values increase memory requests to avoid application crashes.
  memoryRiskPenalty: 50.0
  
  # Multiplier applied if actual OOM events are detected in Prometheus.
  oomMultiplier: 10.0
  
  # Safety buffer applied to predicted peak usage (1.2 = 20% buffer)
  bufferMultiplier: 1.2
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

## 🧮 Algorithms & Calculations

### Resource Prediction Algorithms

#### 1. **EWMA (Exponential Weighted Moving Average)**
Smooths time-series data by giving more weight to recent observations.

**Formula:**
```
EWMA(t) = α × Value(t) + (1 - α) × EWMA(t-1)
where α = smoothing factor (default: 0.3)
```

**Example:**
```
Data: [100, 120, 110, 130, 125]
α = 0.3

EWMA(1) = 100
EWMA(2) = 0.3 × 120 + 0.7 × 100 = 106
EWMA(3) = 0.3 × 110 + 0.7 × 106 = 107.2
EWMA(4) = 0.3 × 130 + 0.7 × 107.2 = 114.04
EWMA(5) = 0.3 × 125 + 0.7 × 114.04 = 117.33

Result: 117.33 (recommended CPU/Memory)
```

#### 2. **Linear Regression**
Predicts future resource needs based on historical trends.

**Formula:**
```
y = mx + b
where:
  m = slope = Σ((x - x̄)(y - ȳ)) / Σ((x - x̄)²)
  b = intercept = ȳ - m × x̄
```

**Example:**
```
Time (hours): [1, 2, 3, 4, 5]
CPU Usage:    [50, 55, 60, 65, 70]

x̄ = 3, ȳ = 60
m = 5 (CPU increases 5% per hour)
b = 45

Prediction for hour 6: y = 5 × 6 + 45 = 75%
```

#### 3. **Percentile Analysis**
Identifies resource usage at specific percentiles to handle spikes.

**Formula:**
```
P(k) = value at position: k × (n + 1) / 100
where k = percentile (50, 95, 99), n = data points
```

**Example:**
```
Sorted CPU values: [10, 20, 30, 40, 50, 60, 70, 80, 90, 100]

P50 (median) = 55 (typical usage)
P95 = 95 (handles 95% of cases)
P99 = 99 (handles spikes)

Recommendation: Set limit at P95 + 20% buffer = 114
```

### AI Anomaly Detection Algorithms

#### 1. **Modified Z-Score**
Detects outliers using median absolute deviation (robust to outliers).

**Formula:**
```
Modified Z-Score = 0.6745 × (x - median) / MAD
where MAD = median(|x - median|)

Anomaly if |Modified Z-Score| > threshold (default: 3.5)
```

**Example:**
```
CPU values: [50, 52, 51, 53, 150, 52, 51]
Median = 52
MAD = median([2, 0, 1, 1, 98, 0, 1]) = 1

For value 150:
Z = 0.6745 × (150 - 52) / 1 = 66.08 > 3.5 ✓ ANOMALY
```

#### 2. **IQR (Interquartile Range)**
Identifies outliers based on quartile distribution.

**Formula:**
```
IQR = Q3 - Q1
Lower bound = Q1 - 1.5 × IQR
Upper bound = Q3 + 1.5 × IQR

Anomaly if value < lower bound OR value > upper bound
```

**Example:**
```
Sorted values: [10, 20, 30, 40, 50, 60, 70, 80, 90, 200]

Q1 = 30, Q3 = 80
IQR = 80 - 30 = 50
Upper bound = 80 + 1.5 × 50 = 155

Value 200 > 155 ✓ ANOMALY
```

#### 3. **Autocorrelation (Pattern Detection)**
Measures correlation between time series and its lagged version.

**Formula:**
```
r(k) = Σ((x(t) - x̄)(x(t-k) - x̄)) / Σ((x(t) - x̄)²)
where k = lag (hours/days)

Pattern detected if r(k) > 0.7
```

**Example:**
```
Hourly CPU: [50, 60, 70, 50, 60, 70, 50, 60, 70]
Lag 3: r(3) = 0.95 ✓ 3-hour pattern detected
```

### Workload DNA Metrics

#### 1. **Volatility (Coefficient of Variation)**
```
CV = (σ / μ) × 100%
where σ = standard deviation, μ = mean

Low volatility: CV < 20% (stable)
High volatility: CV > 50% (unpredictable)
```

**Example:**
```
CPU values: [45, 50, 55, 50, 45]
Mean = 49, StdDev = 4.18
CV = (4.18 / 49) × 100 = 8.5% → Stable workload
```

#### 2. **Growth Rate (Linear Regression Slope)**
```
Growth Rate = slope × 100 / mean
Positive: increasing trend
Negative: decreasing trend
```

**Example:**
```
Daily CPU: [100, 105, 110, 115, 120]
Slope = 5 MB/day
Mean = 110
Growth Rate = (5 / 110) × 100 = 4.5% per day
```

#### 3. **Predictability (R-squared)**
```
R² = 1 - (SS_residual / SS_total)
where:
  SS_residual = Σ(y - ŷ)²
  SS_total = Σ(y - ȳ)²

R² > 0.8: Highly predictable
R² < 0.5: Unpredictable
```

**Example:**
```
Actual: [100, 110, 120, 130, 140]
Predicted: [98, 108, 118, 128, 138]

SS_residual = 20
SS_total = 1000
R² = 1 - (20/1000) = 0.98 → Highly predictable
```

### Scoring System

**Overall Score Calculation:**
```
Score = (HPA_Score × 0.35) + (VPA_Score × 0.35) + (KEDA_Score × 0.30)

Component scores (0-10):
- Resource efficiency: (1 - |actual - recommended| / actual) × 10
- Scaling responsiveness: (1 - avg_lag / target_lag) × 10
- Configuration accuracy: matches / total_checks × 10

Multi-component bonus: +0.5 if 2+ autoscalers configured
```

**Example:**
```
HPA Score: 8.5 (good scaling)
VPA Score: 7.0 (moderate resource fit)
KEDA Score: 9.0 (excellent event-driven scaling)

Overall = (8.5 × 0.35) + (7.0 × 0.35) + (9.0 × 0.30)
        = 2.975 + 2.45 + 2.7
        = 8.125 → Grade: A
```

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

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:
- Setting up your development environment
- Code style guidelines
- Submitting pull requests
- Running tests

## 📋 Prerequisites

- **Go**: 1.21 or higher
- **Kubernetes**: 1.24+ cluster access
- **Prometheus**: Running instance with metrics
- **kubectl**: Configured with cluster access

## 🧪 Testing

```bash
# Run all tests
make test

# Run AI-specific tests
make test-ai

# Run example test scenarios
make test-examples

# Run integration tests
make test-integration

# Run tests with coverage
make coverage

# Run linters
make lint

# Run benchmarks
make bench
```

### Example Test Scenarios

We provide comprehensive test scenarios with sample data:

```bash
# Run all example scenarios
make test-examples
```

**Included Scenarios:**
1. ✅ Steady Workload (no anomalies expected)
2. ✅ Single Spike Detection (1 critical anomaly)
3. ✅ Multiple Spikes (3 anomalies)
4. ✅ Pattern Shift Detection (behavioral change)
5. ✅ Real-World Web Service (daily patterns + traffic spike)

See [docs/EXAMPLES.md](docs/EXAMPLES.md) for detailed examples with expected outputs.

## 🐛 Troubleshooting

### Common Issues

**"failed to connect to Prometheus"**
- Verify Prometheus URL is accessible
- Check network connectivity
- Ensure Prometheus is running

**"workload not found"**
- Verify namespace is correct
- Check workload name spelling
- Ensure kubectl can access the cluster

**"insufficient data"**
- Increase time window with `--time-window`
- Verify Prometheus has historical data
- Check metric retention settings

For more help, see [issues](https://github.com/vasudevchavan/kubecaps/issues) or open a new one.

## 📚 Documentation

- **[AI Examples & Test Cases](docs/EXAMPLES.md)** - Comprehensive examples with sample data and expected outputs
- **[AI Capabilities Roadmap](docs/AI_CAPABILITIES_ROADMAP.md)** - Future AI enhancements (LSTM, RL, NLP, etc.)
- **[Contributing Guide](CONTRIBUTING.md)** - Development setup and guidelines
- **[Changelog](CHANGELOG.md)** - Version history and updates
- [Architecture Overview](docs/ARCHITECTURE.md) _(coming soon)_
- [Algorithm Details](docs/ALGORITHMS.md) _(coming soon)_
- [API Reference](https://pkg.go.dev/github.com/vasudevchavan/kubecaps)

## 🗺️ Roadmap

### ✅ Completed (v0.2.0)
- [x] AI-powered anomaly detection (4 statistical methods)
- [x] Workload DNA behavioral profiling
- [x] Configurable optimization engine via YAML
- [x] Pattern recognition (hourly/daily/weekly)
- [x] Comprehensive test suite with examples
- [x] CI/CD pipeline with GitHub Actions
- [x] Code quality tools (linting, coverage)

### 🚧 In Progress
- [ ] Support for custom metrics
- [ ] Web UI dashboard
- [ ] Historical trend analysis
- [ ] Cost optimization recommendations

### 🔮 Future (See [AI Roadmap](docs/AI_CAPABILITIES_ROADMAP.md))
- [ ] LSTM-based forecasting
- [ ] Reinforcement learning for autoscaling
- [ ] NLP-powered log analysis
- [ ] Graph Neural Networks for cluster topology
- [ ] Multi-cluster support
- [ ] Slack/Teams notifications
- [ ] GitOps integration

## 🛡 License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [client-go](https://github.com/kubernetes/client-go) - Kubernetes client
- [Prometheus Go client](https://github.com/prometheus/client_golang) - Metrics collection

## 📧 Contact

- **Issues**: [GitHub Issues](https://github.com/vasudevchavan/kubecaps/issues)
- **Discussions**: [GitHub Discussions](https://github.com/vasudevchavan/kubecaps/discussions)

---

**⭐ If you find KubeCaps useful, please consider giving it a star!**