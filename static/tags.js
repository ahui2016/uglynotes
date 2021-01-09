ajaxGet('/api/tag/all', null, that => {
  if (!that.response) {
    $('#count').text(0);
    return;
  }
  $('#count').text(that.response.length);

  that.response.forEach(tag => {
    addTagItem(tag, '#sort-by-name>div');
  });
}, () => {
  // onloadend
  $('#loading').hide();
});

ajaxGet('/api/tag/all-by-date', null, that => {
  that.response.forEach(tag => {
    addTagItem(tag, '#sort-by-date>div');
  });
});

function addTagItem(tag, insertPoint) {
  const createdAt = dayjs(tag.CreatedAt);
  const item = $('#li-tmpl').contents().clone();
  item.insertBefore(insertPoint);

  item.find('.datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
  item.find('.name')
    .attr('href', '/html/tag/?name=' + encodeURIComponent(tag.Name))
    .text(tag.Name);
  
  const count = tag.NoteIDs.length;
  item.find('.count').text(count);
  if (count == 0) {
    const delBtnBlock = $('#del-btn-tmpl').contents().clone();
    delBtnBlock.appendTo(item);
  }
}

$('input[name="sort-by"]').change(() => {
  $('#sort-by-name').toggle();
  $('#sort-by-date').toggle();
});
