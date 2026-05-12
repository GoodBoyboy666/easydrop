/**
 * EasyDrop 说说嵌入脚本
 *
 * 将 EasyDrop 的说说列表嵌入到任意第三方页面。
 * 本文件由用户自行部署到自己的服务器或 CDN，不依赖 EasyDrop 服务端托管。
 *
 * === 使用方式 ===
 *
 * <div id="easydrop-talk"></div>
 * <script src="/your-path/talk.js"
 *         data-easydrop-base="https://your-easydrop.example.com"
 *         data-count="5"
 *         data-container="easydrop-talk">
 * </script>
 *
 * 也可以通过 JS 全局变量配置（PJAX 站点必须使用此方式）：
 * <script>
 * window.EasyDropTalk = {
 *   baseUrl: 'https://your-easydrop.example.com',
 *   count: 5,
 *   container: 'easydrop-talk'
 * };
 * </script>
 *
 * 注意：PJAX / Turbolinks 等动态加载站点，data-* 属性可能失效，
 * 请务必使用 window.EasyDropTalk 全局变量方式配置。
 * data-* 属性优先级高于 window.EasyDropTalk。
 *
 * === 配置项 ===
 *
 * data-easydrop-base   EasyDrop 主站地址（必填，不含末尾斜杠）
 * data-count           拉取说说数量，范围 1-100，默认 5
 * data-container       渲染目标容器 ID，默认 easydrop-talk
 *
 * === 依赖 ===
 *
 * 无外部依赖。仅需 EasyDrop 服务端配置 CORS 允许跨域访问。
 */

(function () {
  'use strict';

  var DEFAULT_CONTAINER_ID = 'easydrop-talk';
  var DEFAULT_COUNT = 5;
  var MAX_COUNT = 100;
  var MIN_COUNT = 1;

  var script = document.currentScript ||
                document.querySelector('script[data-easydrop-base]');
  var globalConfig = window.EasyDropTalk || {};

  function getAttr(name, defaultValue, transform, globalName) {
    var attrVal = script ? script.getAttribute('data-' + name) : null;
    var preferredGlobalName = globalName || name;
    var globalVal = globalConfig[preferredGlobalName];

    if (globalVal === undefined && preferredGlobalName !== name) {
      globalVal = globalConfig[name];
    }

    var val = attrVal !== null ? attrVal : (globalVal !== undefined ? globalVal : defaultValue);
    return transform ? transform(val) : val;
  }

  var BASE_URL = getAttr('easydrop-base', null, function (v) {
    if (v === null || v === undefined || String(v).trim() === '') {
      console.error('[EasyDrop Talk] 缺少必填配置 data-easydrop-base（或 window.EasyDropTalk.baseUrl）。' +
                    'PJAX 站点请务必使用 window.EasyDropTalk 全局变量方式配置');
      return null;
    }

    v = String(v).trim().replace(/\/+$/, '');
    try {
      var parsed = new URL(v);
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        throw new Error('仅允许 http/https 协议');
      }
    } catch (e) {
      console.error('[EasyDrop Talk] data-easydrop-base 不合法:', e.message);
      return null;
    }
    return v;
  }, 'baseUrl');

  var CONTAINER_ID = getAttr('container', DEFAULT_CONTAINER_ID);

  var COUNT = getAttr('count', DEFAULT_COUNT, function (v) {
    var n = parseInt(v, 10);
    if (isNaN(n) || n < MIN_COUNT) return DEFAULT_COUNT;
    if (n > MAX_COUNT) return MAX_COUNT;
    return n;
  });

  /* ========== 内联样式 ========== */
  var CSS = [
    '.easydrop-talk-widget {',
    '  max-width: 680px;',
    '  margin: 0 auto;',
    '  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC",',
    '               "Hiragino Sans GB", "Microsoft YaHei", "Helvetica Neue", Arial, sans-serif;',
    '  font-size: 15px;',
    '  line-height: 1.7;',
    '  color: #333;',
    '}',
    '.easydrop-talk-item {',
    '  padding: 6px 10px;',
    '}',
    '.easydrop-talk-content {',
    '  display: -webkit-box;',
    '  -webkit-line-clamp: 2;',
    '  -webkit-box-orient: vertical;',
    '  overflow: hidden;',
    '  white-space: pre-line;',
    '  word-break: break-word;',
    '}',
    '.easydrop-talk-content a {',
    '  color: #222;',
    '  text-decoration: none;',
    '}',
    '.easydrop-talk-content a:hover {',
    '  color: #49b1f5;',
    '}',
    '.easydrop-talk-date {',
    '  font-size: 12px;',
    '  color: #858585;',
    '}',
    '.easydrop-talk-skeleton {',
    '  padding: 8px 10px;',
    '}',
    '.easydrop-talk-skeleton-line {',
    '  height: 14px;',
    '  background: linear-gradient(90deg, #eee 25%, #f5f5f5 50%, #eee 75%);',
    '  background-size: 200% 100%;',
    '  animation: easydrop-talk-shimmer 1.5s infinite;',
    '  border-radius: 4px;',
    '  margin-bottom: 6px;',
    '}',
    '.easydrop-talk-skeleton-line:last-child {',
    '  width: 60%;',
    '  margin-bottom: 0;',
    '}',
    '@keyframes easydrop-talk-shimmer {',
    '  0% { background-position: 200% 0; }',
    '  100% { background-position: -200% 0; }',
    '}',
    '.easydrop-talk-error,',
    '.easydrop-talk-empty {',
    '  padding: 24px 16px;',
    '  color: #999;',
    '  text-align: center;',
    '  font-size: 13px;',
    '}',
    '.easydrop-talk-powered {',
    '  text-align: right;',
    '  padding: 10px 16px 0;',
    '  font-size: 11px;',
    '  color: #bbb;',
    '}',
    '.easydrop-talk-powered a {',
    '  color: #bbb;',
    '  text-decoration: none;',
    '}',
  ].join('\n');

  /* ========== 工具函数 ========== */
  function escapeHtml(str) {
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function stripMarkdown(text) {
    if (!text) return '';

    return text
      .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
      .replace(/\[([^\]]*)\]\([^)]+\)/g, '$1')
      .replace(/^#{1,6}\s+/gm, '')
      .replace(/(\*{1,3}|_{1,3})(.*?)\1/g, '$2')
      .replace(/~~(.*?)~~/g, '$1')
      .replace(/```[\s\S]*?```/g, '')
      .replace(/`([^`]*)`/g, '$1')
      .replace(/^>\s?/gm, '')
      .replace(/^[-*_]{3,}\s*$/gm, '')
      .replace(/^[\s]*[-*+]\s+/gm, '')
      .replace(/^[\s]*\d+\.\s+/gm, '')
      .replace(/<\/?[^>]+(>|$)/g, '')
      .replace(/\n{3,}/g, '\n\n')
      .trim();
  }

  function formatDate(dateStr) {
    var date = new Date(dateStr);
    if (isNaN(date.getTime())) return '';

    var y = date.getFullYear();
    var m = ('0' + (date.getMonth() + 1)).slice(-2);
    var d = ('0' + date.getDate()).slice(-2);
    return y + '-' + m + '-' + d;
  }

  /* ========== 渲染 ========== */
  function renderSkeleton(container) {
    var html = '';
    for (var i = 0; i < COUNT; i++) {
      html += '<div class="easydrop-talk-skeleton">';
      html += '<div class="easydrop-talk-skeleton-line"></div>';
      html += '<div class="easydrop-talk-skeleton-line"></div>';
      html += '</div>';
    }
    container.innerHTML = html;
  }

  function renderItems(container, items) {
    if (!items || items.length === 0) {
      container.innerHTML = '<div class="easydrop-talk-empty">暂无说说</div>';
      return;
    }

    var html = '';
    for (var i = 0; i < items.length; i++) {
      var post = items[i];
      var content = stripMarkdown(post.content);
      var date = formatDate(post.created_at);

      html += '<div class="easydrop-talk-item">';
      var postUrl = BASE_URL + '/posts/' + post.id;
      html += '<div class="easydrop-talk-content"><a href="' + escapeHtml(postUrl) + '" target="_blank" rel="noopener noreferrer">' + escapeHtml(content) + '</a></div>';
      html += '<div class="easydrop-talk-date">' + escapeHtml(date) + '</div>';
      html += '</div>';
    }

    html += '<div class="easydrop-talk-powered">';
    html += 'Powered by <a href="https://github.com/GoodBoyboy666/easydrop" target="_blank" rel="noopener noreferrer">EasyDrop</a>';
    html += '</div>';

    container.innerHTML = html;
  }

  function renderError(container, message) {
    container.innerHTML = '<div class="easydrop-talk-error">' +
      escapeHtml(message || '加载失败，请稍后重试') + '</div>';
  }

  /* ========== 启动 ========== */
  function boot() {
    if (!BASE_URL) {
      console.error('[EasyDrop Talk] 未提供有效的 data-easydrop-base');
      return;
    }

    var container = document.getElementById(CONTAINER_ID);
    if (!container) {
      console.warn('[EasyDrop Talk] 未找到容器元素 #' + CONTAINER_ID);
      return;
    }

    if (container.classList.contains('easydrop-talk-widget')) {
      return;
    }

    container.classList.add('easydrop-talk-widget');

    if (!document.getElementById('easydrop-talk-css')) {
      var style = document.createElement('style');
      style.id = 'easydrop-talk-css';
      style.textContent = CSS;
      document.head.appendChild(style);
    }

    renderSkeleton(container);

    var url = BASE_URL + '/api/v1/posts?size=' + COUNT + '&page=1&order=created_at_desc';

    fetch(url)
      .then(function (res) {
        if (!res.ok) throw new Error('HTTP ' + res.status);
        return res.json();
      })
      .then(function (data) {
        var pinnedItems = Array.isArray(data && data.pinned_items) ? data.pinned_items : [];
        var items = Array.isArray(data && data.items) ? data.items : [];
        var mergedItems = pinnedItems.concat(items).slice(0, COUNT);
        renderItems(container, mergedItems);
      })
      .catch(function (err) {
        console.error('[EasyDrop Talk] 请求失败:', err);
        renderError(container, '加载失败，请稍后重试');
      });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }
})();
