<div align="center">

![new-api](/web/public/logo.png)

# BetterNew-Api

🍥 **基于 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 的定制分支**

<p align="center">
  <strong>简体中文</strong> |
  <a href="./README.en.md">English</a>
</p>

<p align="center">
  <a href="https://raw.githubusercontent.com/QuantumNous/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/QuantumNous/new-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/Likhixang/BetterNew-Api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Likhixang/BetterNew-Api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/QuantumNous/new-api">
    <img src="https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-blue" alt="upstream">
  </a>
</p>

</div>

## 项目简介

本项目为 [QuantumNous/new-api](https://github.com/QuantumNous/new-api)（下一代 LLM 网关与 AI 资产管理系统）的 Fork 定制版，在保留上游全部功能的基础上，针对渠道管理与模型维护流程进行了优化。

- **Fork 来源**：[QuantumNous/new-api](https://github.com/QuantumNous/new-api)
- **开源协议**：AGPL-3.0（保留上游版权声明与许可证文本）

## 与上游的差异

### 一、渠道密钥管理

- **密钥内联展示**：密钥直接内嵌于输入框，默认隐藏，可切换明文显示，并支持一键复制
- **密钥脱敏预览**：隐藏状态下，每行仅显示前 4 位与后 4 位字符（`sk-a1b2****9z8y`），不再整框遮蔽为圆点
- **脱敏不阻塞编辑**：隐藏状态下同样可以直接输入/粘贴新密钥，脱敏只影响展示，无需先切换明文
- **所见即所得编辑**：换行新增密钥即追加，全选重写即替换，无需手动选择编辑模式
- **取消密钥查看校验**：查看密钥不再要求输入验证码
- **移除 OpenAI 组织 ID 字段**：渠道配置不再强制填写组织 ID
- **移动端布局适配**：密钥输入框采用固定宽度，内容超长时内部滚动，避免破坏表单布局

### 二、上游模型检测

- **模型白名单（筛选所需模型）**：仅当上游模型匹配白名单时才纳入可添加范围，支持精确名称与 `regex:` 正则表达式
- **模型黑名单（忽略模型）**：排除与黑名单匹配的上游模型
- **组合过滤**：白名单与黑名单可独立配置或同时生效，过滤顺序为先白名单后黑名单
- **默认启用**：上游模型更新检测默认开启

### 三、模型获取对话框

- **全选 / 取消全选**：可批量勾选当前可见模型
- **仅显示未添加**：过滤已添加的模型，聚焦待添加项

### 四、界面与构建

- **界面语言**：仅保留简体中文与英文
- **自动构建**：通过 GitHub Actions 自动构建镜像并推送至 GHCR，推送至 `main` / `likhixang/patch` 分支或打 `v*` tag 时触发

### 五、PWA 支持

- **可安装**：支持添加到主屏幕 / 桌面，独立窗口运行（`standalone`）
- **离线可用**：Service Worker 预缓存全部静态资源，断网时可打开面板；API 请求（`/api`、`/pg`、`/mj`）始终走网络，绝不缓存
- **动态应用名与图标**：manifest 由后端动态生成（`/manifest.webmanifest`），应用名跟随系统设置中的系统名称（与网页标题同源），PWA 图标跟随系统徽标 URL（未设置时用内置图标）；修改后 Chrome/Edge 会自动同步，iOS 已安装的图标需重新添加到主屏幕
- **实现说明**：基于 Google 官方 [workbox-build](https://developer.chrome.com/docs/workbox/) 在构建期生成 Service Worker（`web/scripts/generate-sw.mjs`），图标由 `web/scripts/generate-pwa-assets.mjs` 基于 `logo.png` 生成（sharp）；采用官方库手动集成，非第三方 PWA 插件方案

### 六、令牌级模型映射（Token-level Model Mapping）

- **Key 级模型重定向**：每个 API Key 可配置模型映射（如 `claude-opus-4-8 → deepseek-v4-pro`），请求到达后先按**原始请求模型**做白名单校验，再重定向为映射目标模型用于渠道选择、上游转发与计费
- **可视化编辑**：令牌表单提供 Visual（表格行：原始模型 → 替换模型，下拉选择）与 JSON（1:1 编辑 + 格式化/复制）双视图，支持重复映射拦截

### 七、模型合并（Model Merge）

- **入站模型名归一化**：全局规则将请求中的模型名（精确匹配或 Go 正则）合并为规范目标模型（如 `cx/deepseek-v4-flash` → `deepseek-v4-flash`），作用于渠道选择、上游转发与计费
- **自动渠道解析（merge 反向匹配）**：渠道选择时，除精确模型名外，所有「模型名经合并规则可归一化为目标模型」的渠道自动进入候选池——请求 `longcat-2.0` 可自动路由到只提供 `longcat-2.0-free` 的渠道，**无需在渠道侧配置 model_mapping 或手动注册模型**（对齐 AxonHub 模型关联语义）
- **全候选随机分配**：精确匹配渠道与 merge 匹配渠道**同池参与权重随机**（无精确优先），选中渠道失败自动重试切换其他候选
- **上游自动改名**：命中 merge 匹配渠道时，系统按渠道真实模型名转发（如请求 `longcat-2.0` → 上游收到 `longcat-2.0-free`），渠道显式配置的 model_mapping 优先于自动推导
- **多别名支持**：一条规则可配置多个别名（每行一个），全部独立注册、共享同一目标模型
- **实时匹配预览**：编辑规则时按渠道实时预览命中的模型与渠道，精确匹配优先、正则按规则 ID 升序
- **标准表格框架**：规则列表使用标准 DataTablePage 框架（分页 / 每页条数 / 总条数），抽屉采用标准布局

### 八、全局提示词注入（Global Prompt Injection）

- **全局 System Prompt**：系统设置 → 模型与路由 → 全局模型配置，可为所有转发的请求注入统一的系统提示词
- **三种注入模式**：前置（prepend）/ 追加（append）/ 覆盖（override）现有系统提示词
- **全协议覆盖**：OpenAI、Claude、Gemini 协议均生效；渠道级系统提示词仍可在其上叠加

### 九、令牌级渠道限制（Token-level Channel Limits）

- **Key 级渠道白名单**：每个 API Key 可配置允许使用的渠道列表（如 `Hermes → 8 个渠道`），请求只路由到白名单内的渠道，其余渠道一律拒绝
- **模型 + 渠道双重限制**：渠道限制与模型限制独立配置、同时生效——先按模型白名单校验，再在渠道白名单内做渠道选择
- **空值即全通**：渠道限制留空表示允许所有渠道（与模型限制语义一致）
- **分组机制隐藏**：传统分组（Group）机制在代码层保留（作为渠道候选集来源），但已在界面完全隐藏——API Key 编辑、渠道编辑、渠道列表均不再出现分组痕迹，无需手动管理分组
- **自动禁用可恢复**：渠道因连续失败被系统自动禁用（状态=3）后，管理员可手动恢复启用

### 十、其他改进

- **渠道测试模型下拉**：测试模型改为从渠道已配置模型中选择的真实下拉框（含"Models & Groups"来源），宽度自适应
- **表单重置修复**：创建渠道成功后关闭抽屉会正确重置表单与高级设置面板状态
- **使用日志 IP 列**：通用日志新增 IP 列（个人资料 → 通知偏好开启「Record IP Address」后，消费/错误日志记录客户端 IP 并展示）
- **模型页默认模型合并**：模型页面 Tab 顺序为模型合并 → 元信息 → 部署，默认进入模型合并

## 快速开始

### 镜像

| 用途 | 镜像 |
|---|---|
| **生产镜像**（自动更新） | `ghcr.io/likhixang/betternew-api:latest` |
| **固定版本镜像**（v1.0.0-rc.23-patch3） | `ghcr.io/likhixang/betternew-api:v1.0.0-rc.23-patch3` |
| **测试镜像**（patch/test 分支） | `ghcr.io/likhixang/betternew-api:patch-test` |

```bash
docker run -d --restart always --name betternew-api \
  -p 3000:3000 \
  -v ./data:/data \
  ghcr.io/likhixang/betternew-api:latest
```

或使用仓库内 `docker-compose.yml`：

```bash
docker compose up -d
```

## 文档

完整功能与部署文档请参阅上游项目：[QuantumNous/new-api](https://github.com/QuantumNous/new-api)

## 许可证

[AGPL-3.0](LICENSE) — 本分支为上游 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 的修改版本，保留上游版权声明与许可证文本。
