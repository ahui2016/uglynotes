
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
      view: () => m('li', {class:'LI', key: note.ID}, [
	m('span', {class:'ID_Date'}, `[id:${note.ID}] ${note.UpdatedAt.format('MMM D, HH:mm')}`),
	m('span', {class:'Deleted', style:{display:note.Config.Deleted}}, 'DELETED'),
	m('span', {class:'Buttons', style:{display:note.Config.Buttons}}, [
	  m('button', {class:'SlimButton'}, 'edit'),
	  m('button', {class:'SlimButton', style:{display:note.Config.DeleteBtn}, onclick: noteComp.ShowDelete}, 'delete'),
	  m('span', {class:'ConfirmBlock', style:{display:note.Config.ConfirmBlock}}, [
	    m('span', {class:'ConfirmDelete'}, note.Config.ConfirmMsg),
	    m('button', {class:'SlimButton', onclick: noteComp.DoDelete, disabled: note.Config.Disabled}, 'yes'),
	    m('button', {class:'SlimButton', onclick: noteComp.CancelDelete}, 'no'),
	  ]),
	]),
	m('br'),
	m('a', {class:'TitleLink', style:{display:note.Config.TitleLink}, href: note.href}, note.Title),
	m('span', {class:'Title', style:{display:note.Config.Title}}, note.Title),
	m('br'),
	m('span', {class:'Tags'}, addPrefix(toTagNames(note.Tags), '#')),
	m(note.Alerts),
      ]),
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
	note.Config.Disabled = true;
	const options = note.Deleted ? noteComp.ReallyDeleteOptions() : noteComp.DeleteOptions();
	m.request(options)
	  .then(noteComp.DeleteSuccess)
	  .catch(noteComp.DeleteFail)
	  .finally(noteComp.DeleteFinally);
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
	console.log(resp);
	note.Config.Title = 'inline';
	note.Config.TitleLink = 'none';
	note.Config.Buttons = 'none';
	note.Config.Deleted = 'inline';
	note.Alerts.Clear();
      },
      DeleteFail: function(e) { note.Alerts.InsertRespErr(e); },
      DeleteFinally: function() { note.Config.Disabled = false; },
    };
    return m(noteComp);
  }
}
