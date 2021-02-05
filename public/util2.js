
function $(selectors) { return document.querySelector(selectors) }
function $S(selectors) { return document.querySelectorAll(selectors) }

const Loading = {
  Display: 'block',
  Hide: () => { Loading.Display = 'none'; },
  view: () => m(
    'p',
    {id:"loading", class:"alert-info", style: {display:Loading.Display}},
    'Loading...')
};

const Alerts = {
  Messages: [],
  Max: 5,
  Insert: function(msgType, msg) {
    Alerts.Messages.unshift(
      {Time: dayjs().format('HH:mm:ss'), Type: msgType, Msg:msg});
    if (Alerts.Messages.length > Alerts.Max) {
      Alerts.Messages.pop();
    }
  },
  InsertRespErr: function(e) {
    const err = !e.response ? `${e.code} ${e.message}` : e.response.message;
    Alerts.Insert('danger', err);
  },
  view: () => m(
    'div', {class: 'alerts'}, Alerts.Messages.map(
      item => m('p', {key: item.Time, class:`alert alert-${item.Type}`}, [
	m('span', item.Time),
	m('span', item.Msg),
      ]))
  )
};

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
      m.redraw();
    }
  };
  return [infoIcon, infoMsg];
}

const Notes = {
  List: [],
  view: () => m(
    'ul', List.map(
      note => m('li', {class: 'LI'}, [

      ]))
  ),
  NewNote: function(note) {
    const updatedAt = dayjs(note.UpdatedAt);
    const noteComp = {
    };
    return m('li', {class: 'LI'}, [
      m('span', {class:'ID_Date'}, `[id:${note.ID}] ${updatedAt.format('MMM D, HH:mm')}`),
      
    ])
  }
}
