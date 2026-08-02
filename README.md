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

## 📌 这是什么

本仓库是 [QuantumNous/new-api](https://github.com/QuantumNous/new-api)（下一代 LLM 网关与 AI 资产管理系统）的 **Fork 定制版**，在上游功能基础上针对实际使用体验做了一系列优化。

**Fork 来源**：[QuantumNous/new-api](https://github.com/QuantumNous/new-api)
**上游协议**：AGPL-3.0（本分支同样适用，保留上游版权声明）

## ✨ 与上游的核心差异

### 1. 渠道密钥管理体验

- **密钥内联显示**：密钥直接显示在输入框内，默认隐藏，点击眼睛图标显示，一键复制
- **脱敏显示**：密钥隐藏时每行显示前 4 位 + 后 4 位（`sk-a1b2****9z8y`），非整框圆点
- **所见即所得编辑**：换行添加新密钥 = 追加，全选重写 = 替换，无需模式选择器
- **去掉查看验证**：不再需要验证码即可查看密钥
- **去掉 OpenAI 组织 ID**：渠道配置不再强制填写组织 ID
- **移动端适配**：长密钥不撑破布局，输入框宽度固定、内容滚动

### 2. 上游模型检测增强

- **筛选所需模型（白名单）**：只添加匹配列表的上游模型，支持精确名与 `regex:` 正则
- **忽略模型（黑名单）**：排除匹配列表的模型（上游原有功能）
- **组合使用**：白名单、黑名单可单独或同时配置（先白名单后黑名单）
- **默认开启**：上游模型更新检测默认启用

### 3. 获取模型对话框

- **全选 / 取消全选**：一键批量勾选当前可见模型列表
- **仅显示未添加**：过滤掉已经添加过的模型

### 4. 界面与构建

- **界面语言**：仅保留简体中文与 English
- **自动构建**：GitHub Actions 构建镜像推送至 GHCR，`main` / `likhixang/patch` 分支与 `v*` tag 自动触发

## 🚀 快速开始

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

## 📚 文档

完整功能与部署文档见上游：[QuantumNous/new-api](https://github.com/QuantumNous/new-api)

## 📄 许可证

[AGPL-3.0](LICENSE) — 本分支为上游 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 的修改版本，保留上游版权声明与许可证文本。
