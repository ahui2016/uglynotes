ajaxGet('/api/note/all', null, that => {
  that.response.forEach(note => {
    let updatedAt = dayjs(note.UpdatedAt);
    let item = $('#li-tmpl').contents().clone();
    item.find('.id').text(note.ID);
    item.find('.datetime').text(updatedAt.format('MMM D, HH:mm'));
    item.find('.title')
      .attr('href', '/html/note?id='+note.ID)
      .text(note.Title);
    item.find('.tags').text(addPrefix(note.Tags, '#'));
    item.insertAfter('#li-tmpl');
  });
}, () => {
  // onloadend
  $('#loading').hide();
});

ajaxGet("/api/note/all/size", null, that => {
  const totalSize = that.response.totalSize;
  const capacity = that.response.capacity;
  const used = fileSizeToString(totalSize, 0);
  const available = fileSizeToString(capacity - totalSize, 0);
  $('#notes-size').text(`已用: ${used}, 剩余可用: ${available}`);
});
