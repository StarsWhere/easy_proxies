# Easy Proxies

[English](README.md) | 简体中文

Easy Proxies 是一个基于 sing-box 的代理池管理工具。

目标是把大量上游节点统一成稳定的本地 HTTP/SOCKS5 代理入口，同时支持按节点独立端口访问。

## 当前能力

- 运行模式：`pool`、`multi-port`、`hybrid`。
- 实际构建的上游协议：`vmess`、`vless`、`trojan`、`ss/shadowsocks`、`hysteria2/hy2`、`socks5/socks`、`http/https`、`anytls`、`tuic`。
- 节点来源：
  - `config.yaml` 的 `nodes`
  - `nodes_file`（每行一个 URI）
  - `subscriptions`（支持 Base64/纯文本/Clash YAML 解析）
- 自动健康检查、失败熔断和黑名单恢复。
- Web 管理面板 + API：
  - 节点状态/探测/导出
  - **手动拉黑/解封节点**
  - 动态设置（`external_ip`、`probe_target`、`skip_cert_verify`、`geoip`）
  - 节点配置增删改查 + 重载
  - 订阅状态查询 + 手动刷新 + **保存即时生效**
  - **实时日志控制台**（最近 1000 行，WebSocket 流式传输）
- 新增可配置 DNS 解析器（对 VMess 域名节点非常关键）。
- 可选 GeoIP 标记（支持 JP/KR/US/HK/TW/SG 地域分区，可在 WebUI 中开关，支持自动更新和热重载）。
- **可配置日志轮转**，支持大小限制、备份数量和压缩。

## 快速开始

### 1）准备配置

```bash
cp config.example.yaml config.yaml
cp nodes.example nodes.txt
```

编辑 `config.yaml`，并配置节点来源（`nodes.txt` / `subscriptions` / `nodes`）。

### 2）启动

Docker：

```bash
./start.sh
# 或
docker compose up -d
```

本地运行：

```bash
go run ./cmd/easy_proxies -config config.yaml
```

## 最小配置示例（Pool）

```yaml
mode: pool

listener:
  address: 0.0.0.0
  port: 2323
  username: user
  password: pass

pool:
  mode: sequential    # sequential / random / balance / latency
  failure_threshold: 3
  blacklist_duration: 24h

management:
  enabled: true
  listen: 0.0.0.0:9091
  probe_target: http://cp.cloudflare.com/generate_204
  password: ""

dns:
  server: 223.5.5.5
  port: 53
  strategy: prefer_ipv4

nodes_file: nodes.txt
```

## DNS 配置说明

`dns` 会同时影响 sing-box DNS 客户端和 VMess 域名拨号解析：

```yaml
dns:
  server: 223.5.5.5
  fallback_servers:    # 备用 DNS 服务器（主 DNS 解析失败时使用）
    - 8.8.8.8
    - 1.1.1.1
  port: 53
  strategy: prefer_ipv4
```

`strategy` 可选值：

- `as_is`
- `prefer_ipv4`
- `prefer_ipv6`
- `ipv4_only`
- `ipv6_only`

如果日志中出现 `lookup <domain>: empty result`，请优先检查该 DNS 配置是否可达且策略合理。

## 运行模式

- `pool`：所有节点共享一个本地 HTTP/SOCKS5 入口。
- `multi-port`：每个节点一个独立本地 HTTP/SOCKS5 端口。
- `hybrid`：同时启用 pool + multi-port。

## 节点来源行为

- 配置了 `subscriptions` 时：
  - 会抓取订阅节点并追加到运行节点列表
  - `nodes_file` 作为订阅节点写入路径
  - 启动阶段不再从 `nodes_file` 读取节点
- `nodes`（内联节点）只要存在就会参与运行。

## 协议支持注意事项

运行时真正支持的协议：

- `vmess`
- `vless`
- `trojan`
- `ss` / `shadowsocks`
- `hysteria2` / `hy2`
- `socks5` / `socks`
- `http` / `https`
- `anytls`
- `tuic`

订阅解析阶段可能识别到更多 URI 前缀（兼容输入），但不在上述列表中的协议会在构建阶段被跳过。

## 管理 API（核心）

- `POST /api/auth`
- `GET|PUT /api/settings`
- `GET /api/nodes`
- `POST /api/nodes/{tag}/probe`
- `POST /api/nodes/{tag}/release`
- `POST /api/nodes/{tag}/blacklist`
- `POST /api/nodes/probe-all`（SSE）
- `GET /api/export`
- `GET|PUT /api/subscription/config`
- `GET|POST /api/subscription/status|refresh`
- `GET|POST|PUT|DELETE /api/nodes/config[...]`
- `POST /api/reload`

`management.password` 为空时，Web/API 不要求登录。

## 重要运行说明

- 重载（`/api/reload` 或订阅刷新）会中断现有连接。
- Settings API 会把配置写回 `config.yaml`；部分设置需要重载后才能完全生效。
- 省略项默认值可在 `internal/config/config.go` 中查看。
- 日志轮转通过 `log` 配置段设置；当 `output: file` 时，日志同时写入控制台和文件，并自动轮转。

## 常见问题

### Docker 权限问题

**问题描述**：使用 `docker-compose.yml` 映射配置目录时，可能遇到 "permission denied" 或 "cannot write to /etc/easy_proxies" 等权限错误。

**原因分析**：容器以非 root 用户运行（docker-compose.yml 中指定 `user: "${UID:-10001}:${GID:-10001}"`），但宿主机挂载目录的所有权可能不匹配。

**解决方案**：

1. **使用提供的 `start.sh` 脚本（推荐）**：
   ```bash
   ./start.sh
   ```
   该脚本会自动：
   - 创建 `data` 和 `logs` 目录
   - 设置正确的权限
   - 传递当前用户的 UID/GID 给 Docker

2. **手动修复权限**：
   ```bash
   mkdir -p data logs
   sudo chown -R $(id -u):$(id -g) data logs
   docker compose up -d
   ```

3. **预先创建配置文件**（备选方法）：
   ```bash
   mkdir -p data
   cp config.example.yaml data/config.yaml
   touch data/nodes.txt
   chown -R $(id -u):$(id -g) data
   docker compose up -d
   ```

**docker run 命令方式**：
```bash
mkdir -p data logs
chmod -R u+w data logs
docker run --user $(id -u):$(id -g) \
  -v $(pwd)/data:/etc/easy_proxies \
  -v $(pwd)/logs:/app/logs \
  --network host \
  ghcr.io/jasonwong1991/easy_proxies:latest
```

### 其他常见问题

- **"配置文件未找到"**：确保挂载目录中存在 `config.yaml` 文件
- **"无法绑定端口"**：检查端口是否被其他服务占用
- **"所有节点健康检查失败"**：验证代理 URI 格式正确，且上游服务器可达

## 更新日志

详见 [CHANGELOG.md](CHANGELOG.md)。

## 开发验证

```bash
go test ./...
```

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=jasonwong1991/easy_proxies&type=Date)](https://star-history.com/#jasonwong1991/easy_proxies&Date)

## 致谢

本项目基于 [sing-box](https://github.com/SagerNet/sing-box) 构建 —— 底层所有协议实现、传输层与拨号逻辑都由 sing-box 提供。特别感谢 SagerNet 团队及所有贡献者的卓越工作。

## 许可证

MIT License

