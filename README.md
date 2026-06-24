# tz

轻量级 Linux 主机监控探针，采集 CPU / 内存 / 磁盘 / 网络指标。资源占用极低（稳态 < 12 MB 内存，< 0.05% CPU）。

## 快速开始

```bash
# 编译
buf generate && go mod tidy && make build

# 启动（serve 模式，Prometheus exporter :9100）
./bin/tz serve
```

```bash
curl http://localhost:9100/metrics
curl http://localhost:9100/health   # → {"status":"healthy"}
curl http://localhost:9100/ready    # → {"ready":true}
```

## Docker Compose（1 server + 3 agent）

```bash
docker-compose up -d
docker attach tz-server            # 查看 TUI，q 退出
```

## 三种模式

```bash
tz serve                            # 独立 Prometheus exporter (:9100)
tz agent --config configs/tz.agent.yaml    # gRPC 客户端，推送指标到 server
tz server --config configs/tz.server.yaml  # gRPC 服务器 + TUI，接收 agent 指标
```

## 远程监控（公网云服务器 → 笔记本）

通过 Tailscale 组网，agent 直连笔记本 server：

```bash
# 笔记本 & 所有云服务器都装 Tailscale
tailscale up
```

```bash
# 笔记本（查看 TUI）
./bin/tz server

# 云服务器 agent 配置  修改tz.agent.yaml
grpc_client:
  server_addr: "100.x.x.x:9090"   # 笔记本的 Tailscale IP
  insecure: true
```

```
云服务器 A (tz agent) ──┐
云服务器 B (tz agent) ──┼── gRPC stream ──→ 笔记本 (tz server + TUI)
云服务器 C (tz agent) ──┘                    100.x.x.x:9090
```

也支持无第三方依赖的 SSH 反向隧道：

```bash
# 笔记本开 server 后，云服务器执行
ssh -R 19090:localhost:9090 user@laptop
# agent 连 localhost:19090
```

## 配置

默认值见 `configs/tz.yaml`，通过环境变量覆盖（前缀 `TZ_`）：

```bash
export TZ_AGENT_HOST_ID="web-01"
export TZ_COLLECTORS_CPU_INTERVAL="5s"
export TZ_LOGGING_LEVEL="debug"
```

## 指标

| 类别 | 指标前缀 | 例 |
|------|---------|----|
| CPU | `tz_cpu_*` | `tz_cpu_usage_percent`, `tz_cpu_logical_count` |
| Memory | `tz_memory_*` | `tz_memory_used_percent`, `tz_memory_swap_used_bytes` |
| Disk | `tz_disk_*` | `tz_disk_used_percent{mountpoint="/"}` |
| Network | `tz_network_*` | `tz_network_bytes_recv_total{interface="eth0"}` |

完整列表见 `/metrics` 端点输出。

## 技术栈

gopsutil · prometheus/client_golang · cobra · viper · logrus · bubbletea · gRPC · Buf

## License

MIT
