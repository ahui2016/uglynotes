
function $(selectors, dom) {
  if (!dom) dom = document;
  return dom.querySelector(selectors);
}

function $S(selectors, dom) {
  if (!dom) dom = document;
  return dom.querySelectorAll(selectors);
}

function $hide(selectors, dom) {
  $(selectors, dom).style.display = 'none';
}

function $show_inline(selectors, dom) {
  $(selectors, dom).style.display = 'inline';
}

function $show_block(selectors, dom) {
  $(selectors, dom).style.display = 'block';
}

function mRequest(options, alerts, config, btnDisabled, onSuccess, onFail, onAlways) {
  if (config && btnDisabled) config[btnDisabled] = true;
  m.request(options)
    .then(onSuccess)
    .catch(e => {
      alerts.InsertRespErr(e);
      if (onFail) onFail();
    })
    .finally(() => {
      if (config && btnDisabled) config[btnDisabled] = false;
      if (onAlways) onAlways();
    });
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

const Loading = {
  Display: 'block',
  Hide: () => { Loading.Display = 'none'; },
  view: () => m(
    'p',
    {id:"loading", class:"alert-info", style: {display:Loading.Display}},
    'Loading...')
};

const Spacer =  m('div', {style:'margin-bottom:2em;'});
const BottomLine = m('div', {style:'margin-top:200px;'});

function CreateAlerts(max) {
  if (!max) max = 5;
  const alerts = {
    Messages: [],
    Max: max,
    Insert: function(msgType, msg) {
      alerts.Messages.unshift(
	{Time: dayjs().format('HH:mm:ss'), Type: msgType, Msg:msg});
      if (alerts.Messages.length > alerts.Max) {
	alerts.Messages.pop();
      }
    },
    InsertRespErr: function(e) {
      const err = !e.response ? `${e.code} ${e.message}` : e.response.message;
      alerts.Insert('danger', err);
    },
    Clear: function() {
      alerts.Messages = [];
    },
    view: () => m(
      'div', {class: 'alerts'}, alerts.Messages.map(
	item => m('p', {key: item.Time, class:`alert alert-${item.Type}`}, [
	  m('span', item.Time),
	  m('span', item.Msg),
	]))
    )
  };
  return alerts;
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
