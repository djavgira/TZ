# tz

轻量级 Linux 主机监控探针，采集 CPU / 内存 / 磁盘 / 网络指标。资源占用极低（稳态 < 12 MB 内存，< 0.05% CPU）。

## 开始

### 面板端（笔记本）

```bash
buf generate && go mod tidy && make build
./bin/tz server
```

### VPS 端（云服务器）

```bash
git clone https://github.com/djavgira/TZ && cd TZ
docker compose -f docker-compose.agents.yml up -d --build
```

> `docker-compose.agents.yml` 中修改 `TZ_GRPC_CLIENT_SERVER_ADDR` 为笔记本 Tailscale IP。

## 远程监控（Tailscale）

```bash
# 双方都装
tailscale up
```

```text
云服务器 (tz agent) ──┐
云服务器 (tz agent) ──┼── gRPC stream ──→ 笔记本 (tz server + TUI)
云服务器 (tz agent) ──┘                    100.x.x.x:9090
```

## 三种模式

```bash
tz server   # gRPC 服务器 + TUI，接收 agent 指标
tz agent    # gRPC 客户端，推送指标到 server
tz serve    # 独立 Prometheus exporter (:9100)
```

## 配置

环境变量覆盖（前缀 `TZ_`）：

```bash
export TZ_AGENT_HOST_ID="web-01"
export TZ_GRPC_CLIENT_SERVER_ADDR="100.x.x.x:9090"
```

## 指标

| 类别 | 前缀 | 例 |
| ---- | ---- | --- |
| CPU | `tz_cpu_*` | `tz_cpu_usage_percent` |
| Memory | `tz_memory_*` | `tz_memory_used_percent` |
| Disk | `tz_disk_*` | `tz_disk_used_percent{mountpoint="/"}` |
| Network | `tz_network_*` | `tz_network_bytes_recv_total{interface="eth0"}` |

## 技术栈

gopsutil · prometheus · cobra · viper · logrus · bubbletea · gRPC · Buf

## License

MIT
