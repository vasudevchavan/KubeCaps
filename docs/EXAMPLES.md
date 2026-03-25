# KubeCaps Examples and Test Cases

This document provides detailed examples of using KubeCaps with sample workloads and expected outputs.

---

## 📋 Table of Contents

1. [Quick Start Example](#quick-start-example)
2. [AI Anomaly Detection Example](#ai-anomaly-detection-example)
3. [Workload DNA Profiling Example](#workload-dna-profiling-example)
4. [Resource Analysis Example](#resource-analysis-example)
5. [Autoscaling Evaluation Example](#autoscaling-evaluation-example)
6. [Test Scenarios](#test-scenarios)

---

## 🚀 Quick Start Example

### Scenario: E-commerce Web Service

**Workload**: `payment-service` - A payment processing microservice
**Characteristics**:
- High traffic during business hours (9 AM - 5 PM)
- Low traffic at night
- Occasional spikes during promotions
- Memory-intensive due to caching

### Step 1: Deploy Sample Workload

```bash
kubectl apply -f samples/hpa-workload.yaml
```

### Step 2: Run AI Analysis

```bash
kubecaps ai payment-service \
  --prometheus-url http://prometheus:9090 \
  --namespace production \
  --time-window 168  # 1 week of data
```

### Expected Output:

```
🤖 AI Analysis for Deployment/payment-service (production)

🔍 Anomaly Detection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 CPU Anomalies Detected: 3

  🔴 [2026-03-24 14:23:15] Sudden spike detected
    Value: 2.45 | Expected: 0.85 | Deviation: 1.60 | Score: 0.87

  🟡 [2026-03-23 15:10:42] Moderate deviation from median
    Value: 1.52 | Expected: 0.85 | Deviation: 0.67 | Score: 0.54

  🟢 [2026-03-22 09:05:18] Outside IQR bounds
    Value: 1.35 | Expected: 0.85 | Deviation: 0.50 | Score: 0.42

✅ No memory anomalies detected

📈 Pattern Shifts Detected: 1

  🟡 [2026-03-23 00:00:00] Pattern shift detected
    Value: 0.95 | Expected: 0.75 | Deviation: 0.20 | Score: 0.35

🧬 Workload DNA Profile
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Workload:        payment-service/production

  📊 Behavioral Characteristics:

     Seasonality:         [daily, weekly]
     Volatility:          0.42 (Medium - Moderate)
     Growth Rate:         +3.2%/week (Growing)
     Predictability:      78% (Highly Predictable)
     Cost Sensitivity:    medium
     Latency Sensitivity: critical

  💻 Resource Profile:

     CPU Intensity:       medium
     Memory Intensity:    high
     CPU/Memory Ratio:    0.45 cores/GB
     Burst Capability:    65% (Medium)

  🌐 Traffic Pattern:

     Type:                periodic
     Peak/Avg Ratio:      2.8x
     Daily Variation:     45%
     Weekly Variation:    38%

  📈 Key Metrics:

     CPU Mean:            0.850 cores
     CPU P95:             1.420 cores
     Memory Mean:         1.85 GB
     Traffic Mean:        125.50 RPS
```

---

## 🔍 AI Anomaly Detection Example

### Test Case 1: Detecting Memory Leak

**Scenario**: Application has a slow memory leak

**Sample Data Pattern**:
```
Time    | Memory (GB)
--------|------------
00:00   | 1.0
04:00   | 1.2
08:00   | 1.5
12:00   | 1.9
16:00   | 2.4  ← Anomaly detected
20:00   | 3.1  ← Critical anomaly
```

**Command**:
```bash
kubecaps ai leaky-app \
  --prometheus-url http://prometheus:9090 \
  --namespace staging \
  --sensitivity 1.0
```

**Expected Detection**:
- ✅ Detects gradual increase as pattern shift
- ✅ Flags 16:00 and 20:00 as anomalies
- ✅ Severity: High → Critical
- ✅ Recommendation: Increase memory limit, investigate leak

---

### Test Case 2: Traffic Spike Detection

**Scenario**: Sudden traffic spike during flash sale

**Sample Data Pattern**:
```
Time    | RPS
--------|-----
13:00   | 100
13:30   | 120
14:00   | 850  ← Spike starts
14:30   | 920  ← Peak
15:00   | 150  ← Back to normal
```

**Command**:
```bash
kubecaps ai web-frontend \
  --prometheus-url http://prometheus:9090 \
  --namespace production \
  --time-window 24
```

**Expected Detection**:
- ✅ Detects 14:00-14:30 as anomalies
- ✅ Identifies sudden spike (>50% change)
- ✅ Severity: Critical
- ✅ Recommendation: Enable HPA, increase maxReplicas

---

## 🧬 Workload DNA Profiling Example

### Test Case 3: Batch Job Characterization

**Scenario**: Nightly ETL job

**Characteristics**:
- Runs every night at 2 AM
- High CPU usage for 2 hours
- Predictable pattern
- Low cost sensitivity (can run slower)

**Command**:
```bash
kubecaps ai etl-processor \
  --prometheus-url http://prometheus:9090 \
  --namespace data-pipeline \
  --time-window 720  # 30 days for pattern detection
```

**Expected DNA Profile**:
```json
{
  "workload_name": "etl-processor",
  "namespace": "data-pipeline",
  "seasonality": ["daily"],
  "volatility": 0.85,
  "growth_rate": 1.2,
  "predictability": 0.92,
  "cost_sensitivity": "low",
  "latency_sensitivity": "low",
  "resource_profile": {
    "cpu_intensity": "high",
    "memory_intensity": "medium",
    "cpu_to_memory_ratio": 1.2,
    "burst_capability": 0.15
  },
  "traffic_pattern": {
    "type": "periodic",
    "peak_to_avg_ratio": 8.5,
    "daily_variation": 0.85,
    "weekly_variation": 0.12
  }
}
```

**Insights**:
- ✅ Highly predictable (92%) - good candidate for scheduled scaling
- ✅ Low latency sensitivity - can use spot instances
- ✅ Daily seasonality - use KEDA cron trigger
- ✅ High CPU intensity - optimize for compute

---

### Test Case 4: Microservice Characterization

**Scenario**: User authentication service

**Characteristics**:
- Steady traffic with business hour peaks
- Low volatility
- Critical latency requirements
- Moderate growth

**Expected DNA Profile**:
```json
{
  "seasonality": ["hourly", "daily", "weekly"],
  "volatility": 0.28,
  "growth_rate": 5.5,
  "predictability": 0.75,
  "cost_sensitivity": "high",
  "latency_sensitivity": "critical",
  "resource_profile": {
    "cpu_intensity": "medium",
    "memory_intensity": "low",
    "burst_capability": 0.45
  },
  "traffic_pattern": {
    "type": "periodic",
    "peak_to_avg_ratio": 2.1
  }
}
```

**Recommendations**:
- ✅ Enable HPA with conservative targets (70%)
- ✅ Set minReplicas=3 for high availability
- ✅ Use VPA in "Off" mode (recommendation only)
- ✅ Monitor closely due to critical latency requirements

---

## 📊 Resource Analysis Example

### Test Case 5: Over-Provisioned Workload

**Current Configuration**:
```yaml
resources:
  requests:
    cpu: 2000m
    memory: 4Gi
  limits:
    cpu: 4000m
    memory: 8Gi
```

**Actual Usage** (P95):
- CPU: 450m
- Memory: 1.2Gi

**Command**:
```bash
kubecaps analyze cache-service \
  --prometheus-url http://prometheus:9090 \
  --namespace production
```

**Expected Output**:
```
━━━ Deployment/cache-service (production) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  📊 Recommended Resources (per replica):

     CPU Request:  500m      (current: 2000m, -75%)
     CPU Limit:    800m      (current: 4000m, -80%)
     Mem Request:  1.5Gi     (current: 4Gi, -62%)
     Mem Limit:    2Gi       (current: 8Gi, -75%)

  🎯 Confidence:   85%
  ⏱  Time Window:  168h

  💡 Analysis & Optimization Strategy:
     Classification: Steady
     [HPA] Disabled (Workload is steady, no high-variance scaling needed)
     [VPA] Set Mode=Off (CPU Buffers = 500m, Mem Buffers = 1.5Gi)

     • CPU request can be reduced by 75% (over-provisioned).
     • Memory request can be reduced by 62% (over-provisioned).

  ⚠️  Risk Profile:
     Level: low
     - Safe to apply gradually
```

**Cost Savings**: ~$450/month (assuming $0.05/core-hour, $0.01/GB-hour)

---

## ⚖️ Autoscaling Evaluation Example

### Test Case 6: HPA Misconfiguration

**Current HPA**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server-hpa
spec:
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
```

**Problem**: Target too high, causing latency spikes

**Command**:
```bash
kubecaps evaluate api-server \
  --prometheus-url http://prometheus:9090 \
  --namespace production \
  --evaluate-autoscaling
```

**Expected Output**:
```
╔══════════════════════════════════════════════════════════════╗
║  KubeCaps Autoscaling Evaluation Report                     ║
╚══════════════════════════════════════════════════════════════╝

  📦 Workload:    Deployment/api-server (production)
  📅 Analyzed:    2026-03-25 18:30:00
  ⏱  Window:      24h

  🏆 Overall Optimization Score:  6.5/10 (C)
     ██████░░░░

  📊 Component Scores:

     HPA     5.5/10  Target too high, scaling lag detected
     VPA     Not configured
     KEDA    Not configured

  🔍 Detected Autoscaling:

     ✓ HPA: api-server-hpa (min=2, max=10, current=3)
     ✗ VPA: Not configured
     ✗ KEDA: Not configured

  💡 Insights:

     🟡 [WARNING] HPA Target Accuracy
        HPA target of 80% is too high. Actual CPU utilization frequently
        exceeds 85%, causing scaling delays.
        Evidence: P95 CPU utilization = 87%
        → Recommendation: Lower target to 65-70%

     🔵 [INFO] Scaling Responsiveness
        Average scale-up time: 45 seconds
        This is within acceptable range but could be improved.
        → Consider: Reduce stabilization window

  ⚠️  Risks:

     ⚠️ [MEDIUM] Scaling Lag
        Detected 3 instances where scaling occurred >30s after load increase
        Impact: Potential latency spikes during traffic bursts
```

---

## 🧪 Test Scenarios

### Running the Test Suite

```bash
# Run all AI tests
make test-ai

# Run specific test
go test -v ./internal/ai -run TestAnomalyDetector_RealWorldScenario

# Run with coverage
make coverage

# Run benchmarks
go test -bench=. ./internal/ai
```

### Test Scenario 1: Steady Workload

**Data**: 100 points, mean=1.0, stddev=0.1
**Expected**: 0 anomalies
**Test**: `TestAnomalyDetector_DetectAnomalies/no_anomalies_in_steady_data`

### Test Scenario 2: Single Spike

**Data**: 100 points with 1 spike at index 50 (10x normal)
**Expected**: 1 critical anomaly
**Test**: `TestAnomalyDetector_DetectAnomalies/detect_single_spike`

### Test Scenario 3: Pattern Shift

**Data**: 200 points, first 100 at mean=1.0, next 100 at mean=2.0
**Expected**: Pattern shift detected around index 100
**Test**: `TestAnomalyDetector_DetectPatternAnomalies`

### Test Scenario 4: Web Service Traffic

**Data**: 288 points (24 hours), simulating daily pattern with 2 PM spike
**Expected**: 
- Daily seasonality detected
- Traffic spike at 2 PM flagged as anomaly
- Predictability score > 0.7
**Test**: `TestAnomalyDetector_RealWorldScenario`

---

## 📈 Performance Benchmarks

### Anomaly Detection Performance

```bash
$ go test -bench=BenchmarkAnomalyDetection ./internal/ai

BenchmarkAnomalyDetection-8   	    5000	    245623 ns/op	   89456 B/op	     234 allocs/op
```

**Interpretation**:
- **245 µs per detection** on 288 data points (24 hours)
- Can process **~4,000 workloads per second**
- Memory efficient: ~90 KB per analysis

---

## 🎓 Best Practices

### 1. Data Collection

**Minimum Requirements**:
- At least 100 data points for anomaly detection
- At least 24 hours for daily pattern detection
- At least 7 days for weekly pattern detection

**Recommended**:
- 30 days of data for comprehensive DNA profiling
- 5-minute granularity for accurate spike detection

### 2. Sensitivity Tuning

**Default (1.0)**: Balanced detection
**High (1.5-2.0)**: More sensitive, catches subtle anomalies
**Low (0.5-0.8)**: Less sensitive, only major anomalies

**When to adjust**:
- Increase for critical services (payment, auth)
- Decrease for batch jobs or non-critical services

### 3. Interpreting Results

**Anomaly Severity**:
- **Critical**: Immediate action required
- **High**: Investigate within hours
- **Medium**: Review during next maintenance
- **Low**: Monitor, may be expected behavior

**DNA Predictability**:
- **>80%**: Highly predictable, good for scheduled scaling
- **60-80%**: Moderately predictable, use adaptive scaling
- **<60%**: Unpredictable, need reactive scaling with buffers

---

## 🔗 Related Documentation

- [AI Capabilities Roadmap](AI_CAPABILITIES_ROADMAP.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Sample Workloads](../samples/README.md)

---

## 💬 Feedback

Found an issue or have a suggestion? [Open an issue](https://github.com/vasudevchavan/kubecaps/issues)