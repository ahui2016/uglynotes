"use strict"

const Spacer = { view: () => $('<div style="margin-bottom: 2em;"></div>') };

const BottomLine = { view: () => $('<div style="margin-top: 200px;"></div>') };

const Loading = {
  view: () => $('<p id="loading" class="alert-info">Loading...</p>'),
  hide: () => { $('#loading').hide(); },
};

function CreateInfoPair(name, msg) {
  const infoMsg = {
    id: `#about-${name}-msg`,
    view: () => $(`<div id="about-${name}-msg" class="InfoMessage" style="display:none">${msg}</div>`),
    toggle: () => { $(infoMsg.id).toggle(); },
  };
  const infoIcon = {
    view: () => $(`<img src="/public/info-circle.svg" class="IconButton" alt="info" title="显示/隐藏说明">`)
      .click(infoMsg.toggle),
  };
  return [infoIcon, infoMsg];
}

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
      alerts.Count = 0;
    },
    view: () => {
      const [vnode, id] = m_id('div');
      vnode.addClass('alerts');
      alerts.ID = id;
      return vnode;
    }
  };
  return alerts;
}

const Notes = {
  id: '#notes',
  view: () => m('ul').attr('id', 'notes'),
  newNote: (note) => {
    const noteComp = {
      id: '',
      deleteYesID: '',
      alerts: CreateAlerts(),
      deleteURL: `/api/note/${note.ID}/deleted`,
      reallyDeleteURL: `/api/note/${note.ID}`,
      view: () => {
	const self = noteComp;
	const [vnode, id] = m_id('li');
	self.id = id;
	self.deleteYesID = id+'DeleteYes';
	vnode.addClass('LI').append([
	  m('span').addClass('ID_Date').text(`[id:${note.ID}] ${dayjs(note.UpdatedAt).format('MMM D, HH:mm')}`),
	  m('span').addClass('Deleted').text('DELETED').css('display', note.Deleted ? 'inline' : 'none'),
	  m('span').addClass('Buttons').append([
	    m('button').text('edit').addClass('SlimButton').click(()=>{window.location = `/light/note/edit?id=${note.ID}`}),
	    m('button').text('delete').addClass('SlimButton DeleteBtn').click(self.showDelete),
	    m('span').addClass('ConfirmBlock').css('display', 'none').append([
	      m('span').addClass('ConfirmDelete').text( note.Deleted ? 'delete this note permanently?' : 'delete this note?'),
	      m('button').text('yes').attr('id', self.deleteYesID.slice(1)).addClass('SlimButton DeleteYes').click(self.executeDelete),
	      m('button').text('no').addClass('SlimButton').click(self.cancelDelete),
	    ]),
	  ]),
	  m('br'),
	  m('a').text(note.Title).addClass('TitleLink').attr('href', `/light/note?id=${note.ID}`),
	  m('span').text(note.Title).addClass('TitleText').css('display', 'none'),
	  m('br'),
	  m('span').addClass('Tags').text(addPrefix(toTagNames(note.Tags), '#')),
	  self.Alerts,
	]);
	return vnode;
      },
      showDelete: () => {
	$(`${noteComp.id} .ConfirmBlock`).show();
	$(`${noteComp.id} .DeleteBtn`).hide();
      },
      cancelDelete: () => {
	$(`${noteComp.id} .ConfirmBlock`).hide();
	$(`${noteComp.id} .DeleteBtn`).show();
	noteComp.alerts.Clear();
      },
      executeDelete: () => {
	const body = new FormData();
	body.append('deleted', true);
	const options = note.Deleted
	      ? {method:'DELETE',url:`/api/note/${note.ID}`,alerts:noteComp.alerts,buttonID:noteComp.deleteYesID}
	      : {method:'PUT',url:`/api/note/${note.ID}/deleted`,body:body,alerts:noteComp.alerts,buttonID:noteComp.deleteYesID};
	ajax(options, () => {
	  $(`${noteComp.id} .TitleLink`).hide();
	  $(`${noteComp.id} .TitleText`).show();
	  $(`${noteComp.id} .Buttons`).hide();
	  $(`${noteComp.id} .Deleted`).show();
	});
      },
    }; // end of noteComp
    return noteComp;
  }, // end of newNote
  append: (note) => {
    const elem = Notes.newNote(note);
    $(Notes.id).append(m(elem));
  },
};

function CreateTag(name) {
  return m('a')
    .addClass('Tag')
    .text(name)
    .attr('href', '/html/search?tags=' + encodeURIComponent(name));
}
