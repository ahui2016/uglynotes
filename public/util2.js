
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
  Messages: {},
  Insert: (msgType, msg) => {
    Alerts.Messages[dayjs().format('HH:mm:ss')] = {Type: msgType, Msg:msg};
  },
  view: () => m(
    'div', Object.entries(Alerts.Messages).map(
      ([dt,msg]) => m('p', {key: dt, class:`alert alert-${msg.Type}`}, [
	m('span', dt),
	m('span', msg.Msg),
      ]))
  )
};
