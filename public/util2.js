"use strict";

// 创建历史版本的间隔时间
const DelayOfAutoUpdate = 1000 * 60 * 5; // 5分钟

// 自动保存（自动更新）次数的上限
const AutoUpdateLimit = 100;

// NoteTitleLimit 限制标题的长度。
// 注意：该限制还需要在 settings.go 中设置（为了做后端限制）
const NoteTitleLimit = 200;

// NoteSizeLimit 限制每篇笔记的体积。
// 注意：该限制还需要在 settings.go 中设置（为了做后端限制）
const NoteSizeLimit = 1 << 19; // 512 KB

// 时区设置，比如，如果是北京时间则应设置为 8。
// 注意：时区设置要与后端的时区设置保持一致。
const TimeZone = 0;

// make a new vnode by name, or return its view.
function m(name) {
  if (jQuery.type(name) == 'string') {
    return $(document.createElement(name));
  }
  return name.view();
}

// cc creates a component with an id.
function cc(name, id, elements) {
  if (!id) id = '' + Math.round(Math.random() * 100000000);
  const vnode = m(name).attr('id', id);
  if (elements) vnode.append(elements);
  return {id: '#'+id, raw_id: id, view: () => vnode};
}

function disable(id) { $(id).prop('disabled', true); }

function enable(id) { $(id).prop('disabled', false); }

// options: method, url, body, alerts, buttonID
function ajax(options, onSuccess, onFail, onAlways) {
  if (options.buttonID) disable(options.buttonID);
  const xhr = new XMLHttpRequest();
  xhr.open(options.method, options.url);
  xhr.onerror = () => {
    window.alert('An error occurred during the transaction');
  };
  xhr.addEventListener('load', function() {
    if (this.status == 200) {
      if (onSuccess) {
	const resp = this.responseText ? JSON.parse(this.responseText) : null;
	onSuccess(resp);
      }
    } else {
      const msg = `${this.status} ${this.responseText}`;
      if (options.alerts) {
	options.alerts.Insert('danger', msg);
      } else {
	console.log(msg);
      }
      if (onFail) onFail(this);
    }
  });
  xhr.addEventListener('loadend', function() {
    if (options.buttonID) enable(options.buttonID);
    if (onAlways) onAlways(this);
  });
  xhr.send(options.body);
}

// 获取地址栏的参数。
function getUrlParam(param) {
  let loc = new URL(document.location);
  return loc.searchParams.get(param);
}

// 把文件大小换算为 KB 或 MB
function fileSizeToString(fileSize, fixed) {
  if (fixed == null) {
    fixed = 2
  }
  const sizeMB = fileSize / 1024 / 1024;
  if (sizeMB < 1) {
    return `${(sizeMB * 1024).toFixed(fixed)} KB`;
  }
  return `${sizeMB.toFixed(fixed)} MB`;
}

function toTagNames(simpleTags) {
  return simpleTags.map(tag => tag.Name)
}

function addPrefix(setOrArr, prefix) {
  if (!setOrArr) return '';
  let arr = Array.from(setOrArr);
  if (!prefix) prefix = '';
  return arr.map(x => prefix + x).join(' ');
}

function tag_replace(tags) {
  return tags.replace(/[#;,，'"/\+\n]/g, ' ').trim();
}

function tagsStringToSet(tags) {
  const trimmed = tag_replace(tags);
  if (trimmed.length == 0) return new Set();
  const arr = trimmed.split(/ +/);
  return new Set(arr);
}

function setsAreEqual(a, b) {
  if (a.size !== b.size) return false;
  for (const item of a) if (!b.has(item)) return false;
  return true;
}

function CreateClipboard(alerts, textFunc) {
  const clipboard = new ClipboardJS('#copy', {text: textFunc});
  clipboard.on('success', () => {
    alerts.Insert('success', '笔记内容已复制到剪贴板');
  });
  clipboard.on('error', e => {
    console.error('Action:', e.action);
    console.error('Trigger:', e.trigger);
    alerts.Insert('danger', '复制失败，详细信息见控制台');
  });
}
