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

## 快速开始

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
