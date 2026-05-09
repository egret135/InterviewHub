import { chromium } from 'playwright';
import { writeFileSync, mkdirSync, existsSync } from 'fs';
import { join, resolve } from 'path';

/**
 * Scrape a DeepSeek chat share page and extract the conversation content.
 *
 * Usage:
 *   node scrape_deepseek.mjs --url https://chat.deepseek.com/share/xxxxx [--outdir ./output]
 *
 * Output (saved to outdir):
 *   - screenshot.png   Full-page screenshot
 *   - content.md        Extracted text in Markdown
 *   - page.html         Full HTML source
 *   - result.json       Structured result { url, timestamp, title, content, contentLength, files }
 */

function parseArgs() {
  const args = process.argv.slice(2);
  const opts = { outdir: '.' };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--url' && args[i + 1]) opts.url = args[++i];
    else if (args[i] === '--outdir' && args[i + 1]) opts.outdir = args[++i];
    else if (args[i] === '--help' || args[i] === '-h') {
      console.log('Usage: node scrape_deepseek.mjs --url <deepseek_share_url> [--outdir <dir>]');
      process.exit(0);
    }
  }
  if (!opts.url) {
    console.error('Error: --url is required');
    process.exit(1);
  }
  return opts;
}

async function extractContent(page) {
  // Strategy 1: .ds-markdown blocks
  const blocks = await page.$$eval('.ds-markdown', els =>
    els.map(el => el.innerText).join('\n\n---\n\n')
  );
  if (blocks.trim()) return blocks;

  // Strategy 2: common chat containers
  const fromContainer = await page.evaluate(() => {
    const selectors = [
      '[class*="chat-container"]',
      '[class*="conversation"]',
      'main',
      '[role="main"]',
      '.share-page',
      '[class*="share"]',
    ];
    for (const sel of selectors) {
      const el = document.querySelector(sel);
      if (el && el.innerText.trim().length > 50) return el.innerText;
    }
    return '';
  });
  if (fromContainer.trim()) return fromContainer;

  // Strategy 3: message bubbles
  const messages = await page.$$eval(
    '[class*="message"], [class*="bubble"], [class*="turn"], [class*="item"]',
    els => els.map(el => el.innerText).join('\n\n---\n\n')
  );
  if (messages.trim()) return messages;

  // Fallback: body text
  return await page.evaluate(() => document.body.innerText);
}

async function main() {
  const { url, outdir } = parseArgs();

  const absOutdir = resolve(outdir);
  if (!existsSync(absOutdir)) mkdirSync(absOutdir, { recursive: true });

  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-blink-features=AutomationControlled'],
  });

  const context = await browser.newContext({
    userAgent:
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36',
    viewport: { width: 1440, height: 900 },
    locale: 'zh-CN',
  });

  await context.addInitScript(() => {
    Object.defineProperty(navigator, 'webdriver', { get: () => false });
  });

  const page = await context.newPage();

  console.log(`加载页面: ${url}`);
  await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });

  try {
    await page.waitForSelector(
      '[class*="message"], [class*="chat"], [class*="conversation"], [class*="dialog"], .ds-markdown',
      { timeout: 15000 }
    );
  } catch {
    console.log('等待内容渲染...');
    await page.waitForTimeout(5000);
  }

  const title = await page.title();

  // Screenshot
  const screenshotPath = join(absOutdir, 'screenshot.png');
  await page.screenshot({ path: screenshotPath, fullPage: true });
  console.log(`截图: ${screenshotPath}`);

  // Extract content
  const content = await extractContent(page);
  console.log(`提取文本: ${content.length} 字符`);

  // HTML
  const html = await page.content();
  const htmlPath = join(absOutdir, 'page.html');
  writeFileSync(htmlPath, html, 'utf-8');
  console.log(`HTML: ${htmlPath}`);

  // Markdown
  const md = `# ${title || 'DeepSeek 聊天分享内容'}

> URL: ${url}
> 采集时间: ${new Date().toISOString()}

---

${content || '(未能提取到文本内容，请查看 page.html)'}
`;
  const mdPath = join(absOutdir, 'content.md');
  writeFileSync(mdPath, md, 'utf-8');
  console.log(`文本: ${mdPath}`);

  // JSON result
  const result = {
    url,
    timestamp: new Date().toISOString(),
    title,
    content,
    contentLength: content.length,
    files: { screenshot: screenshotPath, html: htmlPath, markdown: mdPath },
  };
  const jsonPath = join(absOutdir, 'result.json');
  writeFileSync(jsonPath, JSON.stringify(result, null, 2), 'utf-8');

  // Print preview
  console.log('\n========== 内容预览 ==========');
  console.log(content.substring(0, 2000));
  if (content.length > 2000) {
    console.log(`\n... 完整内容共 ${content.length} 字符，见 ${mdPath}`);
  }

  await browser.close();
  return result;
}

main()
  .then(result => {
    console.log('\n采集完成');
    // Write final JSON to stdout for piping
    console.log(JSON.stringify({ ok: true, files: result.files, contentLength: result.contentLength }));
  })
  .catch(err => {
    console.error('采集失败:', err);
    process.exit(1);
  });
