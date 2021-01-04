ajaxGet('/api/tags/all', null, that => {
  if (!that.response) {
    $('#count').text(0);
    return;
  }
  $('#count').text(that.response.length);

  that.response.forEach(tag => {
    let createdAt = dayjs(tag.CreatedAt);
    let item = $('#li-tmpl').contents().clone();
    item.find('.datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
    item.find('.name')
      .attr('href', '/html/tag/?name=' + encodeURIComponent(tag.Name))
      .text(tag.Name);
    item.find('.count').text(tag.NoteIDs.length)
    item.insertBefore('#sort-by-name>div');
  });
}, () => {
  // onloadend
  $('#loading').hide();
});

ajaxGet('/api/tags/all-by-date', null, that => {
  that.response.forEach(tag => {
    let createdAt = dayjs(tag.CreatedAt);
    let item = $('#li-tmpl').contents().clone();
    item.find('.datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
    item.find('.name')
      .attr('href', '/html/tag/?name=' + encodeURIComponent(tag.Name))
      .text(tag.Name);
    item.find('.count').text(tag.NoteIDs.length)
    item.insertAfter('#sort-by-date>div');
  });
});

$('input[name="sort-by"]').change(() => {
  $('#sort-by-name').toggle();
  $('#sort-by-date').toggle();
});
