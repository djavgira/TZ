# pain_tz

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Docker](https://img.shields.io/badge/image-~12MB-2496ED?logo=docker)](Dockerfile)

轻量级 Linux 主机监控探针 — 采集 CPU、内存、磁盘、网络指标，通过 Prometheus 暴露，部署为单文件静态二进制或容器镜像。

---

## 特性

- **低资源占用** — 稳态 CPU < 0.05%，内存 8–12 MB
- **单文件部署** — 纯 Go 静态编译，`CGO_ENABLED=0`，无任何外部依赖
- **四维采集** — CPU / 内存（含 Swap） / 磁盘（含 IO） / 网络（按接口）
- **即插即用** — 启动即采集，`/metrics` 端点可直接被 Prometheus 抓取
- **健康检查** — 内置 `/health` 和 `/ready` 端点，支持容器就绪探针
- **优雅关闭** — SIGTERM/SIGINT → 停止采集器 → 排空 HTTP Server → 退出
- **双模部署** — systemd 服务（物理机/虚拟机）或 Docker 容器

---

## 快速开始

### 二进制部署

```bash
# 构建
git clone https://github.com/Alice/pain_tz.git
cd pain_tz
go mod tidy
make build

# 运行（使用默认配置）
./bin/pain_tz serve

# 指定配置文件
./bin/pain_tz serve --config /etc/pain_tz/pain_tz.yaml
```

### Docker 部署

```bash
# 构建镜像
make docker-build

# 一行运行（采集宿主机指标）
make docker-run

# docker-compose
docker-compose up -d
```

### 验证

```bash
# 指标端点
curl http://localhost:9100/metrics

# 健康检查
curl http://localhost:9100/health
# → {"status":"healthy"}

# 就绪检查（所有采集器至少完成一次采集后返回 ready）
curl http://localhost:9100/ready
# → {"ready":true}
```

---

## 配置

配置文件为 YAML 格式，通过 `--config` 参数指定。所有配置项均可通过环境变量覆盖（`PAIN_TZ_` 前缀，分隔符用 `_`）。

### 完整配置

```yaml
agent:
  host_id: ""                  # 实例标识，留空则使用主机名

collectors:
  cpu:
    enabled: true
    interval: 10s              # 采样间隔（cpu.Percent 会阻塞此时间）
    per_core: false            # 开启后按核心上报（增加基数）

  memory:
    enabled: true
    interval: 15s
    include_swap: true         # 是否包含 swap 指标

  disk:
    enabled: true
    interval: 30s
    mount_points: []           # 监控的挂载点，空为全部（非虚拟文件系统）
    include_io: true           # 是否上报磁盘 IO 计数器
    fs_type_denylist:
      - tmpfs
      - devtmpfs
      - squashfs
      - overlay

  network:
    enabled: true
    interval: 15s
    interfaces: []             # 监控的接口名，空为全部（除 denylist）
    interface_denylist:
      - lo

server:
  listen_addr: ":9100"
  metrics_path: "/metrics"
  health_path: "/health"
  readiness_path: "/ready"
  shutdown_timeout: 10s
  max_header_bytes: 1048576    # 1 MiB
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 120s

logging:
  level: "info"                # debug / info / warn / error
  format: "json"               # json（结构化）或 text（可读）
  output: "stdout"             # stdout 或 file
  file_path: "/var/log/pain_tz/agent.log"
```

### 环境变量覆盖

所有配置均可用 `PAIN_TZ_` 前缀的环境变量覆盖，嵌套用 `_` 连接：

```bash
export PAIN_TZ_SERVER_LISTEN_ADDR=":9200"
export PAIN_TZ_COLLECTORS_CPU_INTERVAL="5s"
export PAIN_TZ_LOGGING_LEVEL="debug"
export PAIN_TZ_AGENT_HOST_ID="prod-web-01"
```

---

## 指标参考

### CPU — `pain_tz_cpu_*`

| 指标 | 类型 | 说明 |
|---|---|---|
| `pain_tz_cpu_usage_percent` | Gauge | 整体 CPU 使用率 (0–100) |
| `pain_tz_cpu_user_percent` | Gauge | 用户态 CPU 时间占比 |
| `pain_tz_cpu_system_percent` | Gauge | 内核态 CPU 时间占比 |
| `pain_tz_cpu_idle_percent` | Gauge | 空闲 CPU 时间占比 |
| `pain_tz_cpu_iowait_percent` | Gauge | I/O 等待 CPU 时间占比 |
| `pain_tz_cpu_logical_count` | Gauge | 逻辑 CPU 核心数 |

### Memory — `pain_tz_memory_*`

| 指标 | 类型 | 说明 |
|---|---|---|
| `pain_tz_memory_total_bytes` | Gauge | 物理内存总量 |
| `pain_tz_memory_used_bytes` | Gauge | 已用内存 |
| `pain_tz_memory_available_bytes` | Gauge | 可用内存（含可回收） |
| `pain_tz_memory_used_percent` | Gauge | 内存使用率 (0–100) |
| `pain_tz_memory_swap_total_bytes` | Gauge | Swap 总量 |
| `pain_tz_memory_swap_used_bytes` | Gauge | Swap 已用 |
| `pain_tz_memory_swap_used_percent` | Gauge | Swap 使用率 (0–100) |

### Disk — `pain_tz_disk_*`

| 指标 | 类型 | 标签 |
|---|---|---|
| `pain_tz_disk_total_bytes` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_used_bytes` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_free_bytes` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_used_percent` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_read_bytes_total` | Counter | `device` |
| `pain_tz_disk_write_bytes_total` | Counter | `device` |

### Network — `pain_tz_network_*`

| 指标 | 类型 | 标签 |
|---|---|---|
| `pain_tz_network_bytes_sent_total` | Counter | `interface` |
| `pain_tz_network_bytes_recv_total` | Counter | `interface` |
| `pain_tz_network_packets_sent_total` | Counter | `interface` |
| `pain_tz_network_packets_recv_total` | Counter | `interface` |
| `pain_tz_network_errors_sent_total` | Counter | `interface` |
| `pain_tz_network_errors_recv_total` | Counter | `interface` |
| `pain_tz_network_drops_sent_total` | Counter | `interface` |
| `pain_tz_network_drops_recv_total` | Counter | `interface` |

---

## Prometheus 抓取配置

```yaml
scrape_configs:
  - job_name: "pain_tz"
    scrape_interval: 15s
    static_configs:
      - targets: ["localhost:9100"]
        labels:
          host_id: "${HOSTNAME}"
```

---

## 部署

### 方式一：Systemd（物理机 / 虚拟机）

```bash
# 安装二进制 + 配置 + systemd unit
make install

# 手动启停
sudo systemctl start pain_tz
sudo systemctl status pain_tz
sudo journalctl -u pain_tz -f
```

systemd unit 已内置安全加固：`ProtectSystem=strict`、`NoNewPrivileges=yes`、`MemoryMax=50M`、`CPUQuota=5%`。

### 方式二：Docker

```bash
# 构建并运行
make docker-build docker-run

# docker-compose
docker-compose up -d
docker-compose logs -f
```

> **注意**：容器内需挂载宿主机 `/proc` 和 `/sys` 才能获取宿主机层面指标。`make docker-run` 和 `docker-compose.yml` 已自动处理。

### 方式三：Kubernetes DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: pain-tz
  labels:
    app: pain-tz
spec:
  selector:
    matchLabels:
      app: pain-tz
  template:
    metadata:
      labels:
        app: pain-tz
    spec:
      hostNetwork: true
      hostPID: true
      containers:
        - name: pain-tz
          image: pain_tz:latest
          imagePullPolicy: IfNotPresent
          env:
            - name: HOST_PROC
              value: /host/proc
            - name: HOST_SYS
              value: /host/sys
            - name: PAIN_TZ_AGENT_HOST_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          ports:
            - containerPort: 9100
              name: metrics
          volumeMounts:
            - name: proc
              mountPath: /host/proc
              readOnly: true
            - name: sys
              mountPath: /host/sys
              readOnly: true
          readinessProbe:
            httpGet:
              path: /ready
              port: 9100
            initialDelaySeconds: 15
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /health
              port: 9100
            initialDelaySeconds: 30
            periodSeconds: 15
          securityContext:
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            runAsUser: 65532
            capabilities:
              drop: ["ALL"]
      volumes:
        - name: proc
          hostPath:
            path: /proc
        - name: sys
          hostPath:
            path: /sys
```

---

## 构建

```makefile
make build          # 编译 linux/amd64 静态二进制 → bin/pain_tz
make cross-build    # 交叉编译 linux/amd64 + linux/arm64
make test           # 运行测试
make test-cover     # 测试 + 覆盖率报告
make lint           # golangci-lint 检查
make docker-build   # 构建 Docker 镜像 (~12 MB)
make docker-run     # 本地容器运行
```

构建产物是完全静态链接的 ELF 二进制，可复制到任何 Linux x86_64/ARM64 主机直接运行。

---

## 项目结构

```
pain_tz/
├── cmd/pain_tz/main.go                # 入口：Cobra CLI + 信号处理
├── internal/
│   ├── agent/agent.go                 # 编排层：Agent 生命周期管理
│   ├── collector/
│   │   ├── collector.go               # Collector 接口定义
│   │   ├── registry.go                # 采集器注册表（goroutine 管理）
│   │   ├── cpu/cpu.go                 # CPU 采集器
│   │   ├── memory/memory.go           # 内存采集器
│   │   ├── disk/disk.go               # 磁盘采集器
│   │   └── network/network.go         # 网络采集器
│   ├── config/config.go               # Viper 配置加载
│   ├── server/
│   │   ├── server.go                  # HTTP Server（/metrics /health /ready）
│   │   └── middleware.go              # 请求日志 + panic 恢复
│   └── exporter/prometheus.go         # Prometheus 指标定义与注册
├── configs/
│   ├── pain_tz.yaml                   # 默认配置
│   └── pain_tz.container.yaml         # 容器优化配置
├── deployments/
│   └── systemd/pain_tz.service        # systemd unit 文件
├── pkg/version/version.go             # 构建时注入的版本信息
├── Dockerfile                         # 多阶段容器构建
├── docker-compose.yml                 # 本地容器编排
├── Makefile                           # 构建 / 测试 / 部署自动化
└── README.md
```

---

## 技术栈

| 模块 | 库 |
|---|---|
| 系统信息采集 | [shirou/gopsutil v3](https://github.com/shirou/gopsutil) |
| Prometheus 指标 | [prometheus/client_golang](https://github.com/prometheus/client_golang) |
| 配置管理 | [spf13/viper](https://github.com/spf13/viper) |
| 命令行 | [spf13/cobra](https://github.com/spf13/cobra) |
| 结构化日志 | [sirupsen/logrus](https://github.com/sirupsen/logrus) |

---

## 性能基线

在空闲的 4 核 Linux VM 上 24 小时稳态测试：

| 指标 | 值 |
|---|---|
| 常驻内存 (RSS) | 8–12 MB |
| CPU 占用 | < 0.05% of one core |
| 磁盘 I/O | 仅 gopsutil 读取 `/proc`，无磁盘写入 |
| 网络 | 仅 Prometheus scrape 产生的出站流量 |
| 二进制大小 (strip) | ~8 MB |

---

## License

MIT
