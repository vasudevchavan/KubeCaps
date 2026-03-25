# KubeCaps Sample Workloads

This directory contains several Kubernetes manifests designed to test and demonstrate the capabilities of **KubeCaps**.

## Available Samples

### 1. HPA Workload (`hpa-workload.yaml`)
Tests Horizontal Pod Autoscaler detection and responsiveness evaluation.
- **Scenario**: A CPU-intensive busybox deployment with an HPA targeting 50% CPU utilization.
- **Target**: `hpa-worker` (Deployment) and `hpa-worker-scaler` (HPA).

### 2. VPA Workload (`vpa-workload.yaml`)
Tests Vertical Pod Autoscaler detection in recommendation mode.
- **Scenario**: A busybox deployment with a VPA set to `updateMode: "Off"`.
- **Target**: `vpa-worker` (Deployment) and `vpa-worker-scaler` (VPA).

### 3. KEDA Workload (`keda-workload.yaml`)
Tests KEDA ScaledObject detection and trigger evaluation.
- **Scenario**: A deployment with a KEDA `prometheus` trigger.
- **Target**: `keda-worker` (Deployment) and `keda-worker-scaler` (ScaledObject).

### 4. Conflict Scenario (`conflict-workload.yaml`)
Tests the **Intelligence Engine's** ability to detect and resolve autoscaling conflicts.
- **Scenario**: A single deployment (`conflict-worker`) targeted by HPA, VPA (Auto mode), and KEDA (CPU trigger) simultaneously.
- **Expected Result**: KubeCaps should identify the conflict and suggest disabling the redundant scalers.

## Usage

1. **Apply a sample**:
   ```bash
   kubectl apply -f samples/hpa-workload.yaml
   ```

2. **Analyze with KubeCaps**:
   ```bash
   kubecaps analyze --namespace demo --workload hpa-worker
   ```

3. **Evaluate with KubeCaps**:
   ```bash
   kubecaps evaluate --namespace demo --workload hpa-worker
   ```

> [!NOTE]
> These samples use the `demo` namespace. Ensure the namespace exists or modify the manifests before applying.
