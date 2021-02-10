"use strict"

const Spacer = { view: () => $('<div style="margin-bottom: 2em;"></div>') };

const BottomLine = { view: () => $('<div style="margin-top: 200px;"></div>') };

const Loading = {
  view: () => $('<p id="loading" class="alert-info">Loading...</p>'),
  hide: () => { $('#loading').hide(); },
};

function CreateAlerts(max) {
  if (!max) max = 5;
  const alerts = {
    ID: '',
    Count: 0,
    Insert: (msgType, msg) => {
      const elem = $(alerts.ID);
      elem.prepend(
	m('p').addClass(`alert alert-${msgType}`)
	  .append(m('span').text(dayjs().format('HH:mm:ss')))
	  .append(m('span').text(msg))
      );
      alerts.Count++;
      if (alerts.Count > max) {
	$(`${alerts.ID} p:last-of-type`).remove();
      }
    },
    Clear: () => {
      $(alerts.ID).html('');
    },
    view: () => {
      const vnode = m('div').addClass('alerts');
      alerts.ID = random_id(vnode);
      return vnode;
    }
  };
  return alerts;
}
