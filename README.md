# pain_tz

轻量级 Linux 主机监控探针，采集 CPU / 内存 / 磁盘 / 网络指标，通过 Prometheus `/metrics` 暴露。资源占用极低（稳态 < 12 MB 内存，< 0.05% CPU）。

## 快速开始

```bash
docker-compose up -d
```

验证：

```bash
curl http://localhost:9100/metrics
curl http://localhost:9100/health   # → {"status":"healthy"}
curl http://localhost:9100/ready    # → {"ready":true}
```

> `docker-compose.yml` 已自动挂载宿主机 `/proc` 和 `/sys` 并注入 `HOST_PROC` / `HOST_SYS` 环境变量，确保采集宿主机而非容器的资源指标。

## 配置

通过环境变量覆盖默认配置，前缀 `PAIN_TZ_`，嵌套用 `_` 连接：

```bash
# docker-compose.yml 中设置，或直接 export
PAIN_TZ_SERVER_LISTEN_ADDR=":9200"
PAIN_TZ_COLLECTORS_CPU_INTERVAL="5s"
PAIN_TZ_LOGGING_LEVEL="debug"
PAIN_TZ_AGENT_HOST_ID="prod-web-01"
```

完整默认配置见 `configs/pain_tz.container.yaml`。

## 指标

### CPU
| 指标 | 类型 |
|---|---|
| `pain_tz_cpu_usage_percent` | Gauge |
| `pain_tz_cpu_user_percent` | Gauge |
| `pain_tz_cpu_system_percent` | Gauge |
| `pain_tz_cpu_idle_percent` | Gauge |
| `pain_tz_cpu_iowait_percent` | Gauge |
| `pain_tz_cpu_logical_count` | Gauge |

### Memory
| 指标 | 类型 |
|---|---|
| `pain_tz_memory_total_bytes` | Gauge |
| `pain_tz_memory_used_bytes` | Gauge |
| `pain_tz_memory_available_bytes` | Gauge |
| `pain_tz_memory_used_percent` | Gauge |
| `pain_tz_memory_swap_total_bytes` | Gauge |
| `pain_tz_memory_swap_used_bytes` | Gauge |
| `pain_tz_memory_swap_used_percent` | Gauge |

### Disk
| 指标 | 类型 | 标签 |
|---|---|---|
| `pain_tz_disk_total_bytes` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_used_bytes` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_free_bytes` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_used_percent` | Gauge | `mountpoint`, `device`, `fstype` |
| `pain_tz_disk_read_bytes_total` | Counter | `device` |
| `pain_tz_disk_write_bytes_total` | Counter | `device` |

### Network
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

## 项目结构

```
├── cmd/pain_tz/main.go          # 入口
├── internal/
│   ├── agent/agent.go           # 编排层
│   ├── collector/               # 采集器 (cpu/memory/disk/network)
│   ├── config/config.go         # 配置加载
│   ├── server/                  # HTTP Server (/metrics /health /ready)
│   └── exporter/prometheus.go   # Prometheus 指标
├── configs/                     # 默认配置 & 容器配置
├── Dockerfile                   # 多阶段容器构建
├── docker-compose.yml
└── Makefile
```

## 技术栈

[gopsutil v3](https://github.com/shirou/gopsutil) · [prometheus/client_golang](https://github.com/prometheus/client_golang) · [viper](https://github.com/spf13/viper) · [cobra](https://github.com/spf13/cobra) · [logrus](https://github.com/sirupsen/logrus)

## License

MIT
