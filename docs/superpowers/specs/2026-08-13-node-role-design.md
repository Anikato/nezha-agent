# Agent node_role 三角色设计

- 日期：2026-08-13
- 范围：同一套 Nezha gRPC；一份 Agent 源码；现有 X-Panel 版 Dashboard
- 角色声明：方案 A，Agent 配置显式写 `node_role`，Dashboard 不猜测

## 目标

让 Dashboard 能稳定区分三种接入角色，避免 X-Panel 升级入口打到普通节点或 OpenWrt 路由器。

| 角色 | `node_role` | 谁写入 | Dashboard 行为 |
| --- | --- | --- | --- |
| 普通原版 | 不写 / 空 | 上游或未改配置的 Agent | 只出现在服务器列表 |
| X-Panel | `xpanel` | X-Panel 安装包的 `config.yml` | 服务器列表 + X-Panel 页 + 允许升级 |
| OpenWrt | `openwrt` | OpenWrt 包的 `config.yml` | 服务器列表；拒绝 X-Panel 升级 |

代码默认必须是空。`xpanel` / `openwrt` 只出现在对应安装包配置里，不写死在二进制里。

## 不做

- 不新开 gRPC 消息、不改 proto
- 不根据 `host.platform` 或 `xpanel_version` 推断角色
- 不做 OpenWrt 管理页、ipk、procd
- 不在后台配置弹窗里编辑 `node_role`
- 不把角色做成三种互不兼容的 Agent 协议

OpenWrt 打包是后续独立任务；那个包必须写入 `node_role: openwrt`。

## 字段与握手

复用现有 `xpanel_name` 路径。

Agent `config.yml`：

```yaml
node_role: xpanel   # 或 openwrt；普通节点省略此字段
```

- `AgentConfig.NodeRole`：`koanf:"node_role" json:"node_role"`
- 合法值（trim + 小写）只有 `xpanel`、`openwrt`
- 空、缺省、非法值：不发送 metadata
- 握手 metadata 与 `xpanel_name` 相同，带两个别名：`node-role`、`node_role`
- `ApplyConfig` 是 `Clone()` 后再 `json.Unmarshal`。后台配置表单不含此字段，因此远程改配置不会把它抹掉

Dashboard `Server.NodeRole`：`json:"node_role,omitempty"`。空即 generic。`RuntimeCopy` 必须带上。

## 握手规则

Dashboard 在现有 `authHandler.check` 里读取角色，不新增 RPC。

1. 解析 allowlist。非法或空视为「本次未声明」。
2. 新注册：写入解析结果（未声明则为空）。
3. 已存在节点：本次声明了合法角色则覆盖；未声明则保留库里的值。
4. 不改节点名字。`xpanel_name` 仍只用于首次命名。

原版 Agent 不带该字段，永远保持 generic，监控/终端/任务不受影响。

## Dashboard 行为

- 服务器列表：三种都显示。仅 `xpanel` / `openwrt` 打标记，generic 不加。
- `/dashboard/xpanel`：只列出 `node_role == "xpanel"`。
- `POST /api/v1/xpanel/version/refresh` 与 `POST /api/v1/xpanel/upgrade`：非 `xpanel` 直接失败并说明原因，不执行 `/opt/xpanel/xpanel`。
- 版本探测仍只用于版本号，不用于判断身份。

## 迁移

已在线的 X-Panel 节点库里没有 `node_role`，过滤一旦打开，它们会暂时从 X-Panel 页消失。

发布顺序：

1. Agent 认识 `node_role`
2. X-Panel 安装/升级把 `config.yml` 写成 `node_role: xpanel`（只合并这一字段，保留 UUID 等）
3. 节点重连后 Dashboard 写入角色
4. 再打开 X-Panel 页过滤和升级拒绝

不做全库回填。需要立刻入页的节点，等 Agent 带上配置重连即可。

## 测试

- Agent：空值不发 metadata；`xpanel` / `openwrt` 发两个别名；非法值不发；`Clone` / `ApplyConfig` 保留字段
- Dashboard：新注册写入；重连覆盖；未声明不覆盖；升级 API 拒绝 generic 与 openwrt
- 前端：X-Panel 页不含非 xpanel 节点

## 验收

- 原版 Agent 能连，角色为空
- X-Panel 包默认 `xpanel`，出现在 X-Panel 页，升级 API 接受
- 配置为 `openwrt` 的 Agent 能连，出现在服务器列表，X-Panel 升级被拒绝
- 后台改 Agent 配置后 `node_role` 仍在
