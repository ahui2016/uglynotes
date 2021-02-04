
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
