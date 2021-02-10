"use strict";

// 创建历史版本的间隔时间
const DelayOfAutoUpdate = 1000 * 60 * 5 // 5分钟

// 自动保存（自动更新）次数的上限
const AutoUpdateLimit = 100

// NoteTitleLimit 限制标题的长度。
// 注意：该限制还需要在 settings.go 中设置（为了做后端限制）
const NoteTitleLimit = 200

// NoteSizeLimit 限制每篇笔记的体积。
// 注意：该限制还需要在 settings.go 中设置（为了做后端限制）
const NoteSizeLimit = 1 << 19 // 512 KB

// make a new vnode by name, or return its view.
function m(name) {
  if (jQuery.type(name) == 'string') {
    return $(document.createElement(name));
  }
  return name.view();
}

// set a random id to vnode and return the id.
function random_id(vnode) {
  vnode.attr('id', Math.round(Math.random() * 100000000));
  return '#' + vnode.attr('id');
}

// return a new vnode and its id.
function m_id(name) {
  const vnode = m(name);
  const id = random_id(vnode);
  return [vnode, id];
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

function InfoPair(name, msg) {
  const infoMsg = {
    Display: 'none',
    view: () => m(
      'div',
      {id: `about-${name}-info`, class: 'InfoMessage', style: {display: infoMsg.Display}},
      msg
    )
  };
  const infoIcon = {
    view: () => m(
      'img',
      {id: `about-${name}-icon`, src: '/public/info-circle.svg', class: 'IconButton', alt: "info", title: "显示/隐藏说明", onclick: infoIcon.Toggle}
    ),
    Toggle: function() {
      if (infoMsg.Display == 'none') {
	infoMsg.Display = 'block';
      } else {
	infoMsg.Display = 'none';
      }
    }
  };
  return [infoIcon, infoMsg];
}

const Notes = {
  List: [],
  view: () => m(
    'ul', Notes.List.map(Notes.NewNote)
  ),
  NewNote: function(note) {
    const noteComp = {
      view: function() {
	const self = noteComp;
	const cfg = note.Config;
	return m('li', {class:'LI', key: note.ID}, [
	  m('span', {class:'ID_Date'}, `[id:${note.ID}] ${note.UpdatedAt.format('MMM D, HH:mm')}`),
	  note.Deleted ? m('span', {class:'Deleted'}, 'DELETED') : '',
	  !note.Exists ? '' : m('span', {class:'Buttons'}, [
	    note.Deleted ? '' : m('button', {class:'SlimButton', onclick:()=>{window.location = cfg.EditUrl}}, 'edit'),
	    m('button', {class:'SlimButton', style:{display:cfg.DeleteBtn}, onclick: self.ShowDelete}, 'delete'),
	    m('span', {class:'ConfirmBlock', style:{display:cfg.ConfirmBlock}}, [
	      m('span', {class:'ConfirmDelete'}, cfg.ConfirmMsg),
	      m('button', {class:'SlimButton', onclick: self.DoDelete, disabled: cfg.Disabled}, 'yes'),
	      m('button', {class:'SlimButton', onclick: self.CancelDelete}, 'no'),
	    ]),
	  ]),
	  m('br'),
	  note.Exists ? m('a', {class:'TitleLink', href: note.href}, note.Title) : '',
	  note.Exists ? '' : m('span', {class:'Title'}, note.Title),
	  m('br'),
	  m('span', {class:'Tags'}, addPrefix(toTagNames(note.Tags), '#')),
	  m(note.Alerts),
	]);
      },
      ShowDelete: function() {
	note.Config.DeleteBtn = 'none';
	note.Config.ConfirmBlock = 'inline';
      },
      CancelDelete: function() {
	note.Config.DeleteBtn = 'inline';
	note.Config.ConfirmBlock = 'none';
	note.Alerts.Clear();
      },
      DoDelete: function() {
	const self = noteComp;
	const options = note.Deleted ? self.ReallyDeleteOptions() : self.DeleteOptions();
	mRequest(
	  options, note.Alerts, note.Config, 'Disabled', self.DeleteSuccess);
      },
      DeleteOptions: function() {
	const body = new FormData();
	body.append('deleted', true);
	return {method:'PUT', url:note.Config.DeleteUrl, body:body}
      },
      ReallyDeleteOptions: function() {
	return {method:'DELETE', url:note.Config.ReallyDeleteUrl}
      },
      DeleteSuccess: function(resp) {
	note.Deleted = true;
	note.Exists = false;
	note.Alerts.Clear();
      },
    };
    return m(noteComp);
  }
}
