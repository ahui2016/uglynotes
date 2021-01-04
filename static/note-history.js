const id = getUrlParam('id');

$('#note-id')
  .text('id:'+id)
  .attr('href', '/html/note?id='+id);

let history_size = 0;

ajaxGet(`/api/note/${id}/history`, null, that => {
  that.response.forEach(history => {
    history_size += history.Size;

    const item = $('#li-tmpl').contents().clone();
    item.insertAfter('#li-tmpl');

    item.find('.id').text(history.ID);
    if (history.Protected) {
      item.find('.protected').show();
    }
    const createdAt = dayjs(history.CreatedAt);
    item.find('.datetime').text(createdAt.format('MMM D, HH:mm'));
    item.find('.size').text(fileSizeToString(history.Size));
    item.find('.title')
      .text(history.Contents)
      .attr('href', `/html/history?id=${history.ID}&noteid=${id}`);
  });

  $('#history-size')
    .show()
    .text(fileSizeToString(history_size));
}, () => {
  // onloadend
  $('#loading').hide();
});
