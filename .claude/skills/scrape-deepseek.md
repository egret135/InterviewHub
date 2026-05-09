---
name: scrape-deepseek
description: 抓取 DeepSeek 聊天分享页面内容。用法：/scrape-deepseek <url>
---

# 抓取 DeepSeek 聊天分享

调用 Playwright 无头浏览器渲染 DeepSeek 分享页面，提取对话文本、截图和 HTML 源码。

## 用法

用户输入 `/scrape-deepseek <url>` 时执行以下命令：

```bash
node /Users/bailu/CodeProject/InterviewHub/tools/scraper/scrape_deepseek.mjs --url "<url>" --outdir "/Users/bailu/CodeProject/InterviewHub/tools/scraper/deepseek_output"
```

## 输出

采集结果保存在 `tools/scraper/deepseek_output/` 目录：
- `content.md` - 提取的对话文本
- `screenshot.png` - 页面截图
- `page.html` - 完整 HTML 源码
- `result.json` - 结构化结果（含 url、title、content、contentLength、文件路径）

## 步骤

1. 执行上述 node 命令
2. 命令完成后，读取 `tools/scraper/deepseek_output/result.json` 获取采集状态
3. 读取 `tools/scraper/deepseek_output/content.md` 获取完整文本内容
4. 向用户总结页面内容

## 如果采集失败

- 检查 DeepSeek 分享链接是否有效
- 如果返回 Rate Limit，等待几分钟后重试
- 可以尝试读取 `tools/scraper/deepseek_output/page.html` 分析页面状态
