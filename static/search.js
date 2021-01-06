const search_input = $('#search-input');
const search_btn = $('#search-btn');
const loading = $('#loading');
const note_list = $('ul');

let searchFor = 'tags';

search_btn.click(event => {
  event.preventDefault();
  const pattern = search_input.val().trim();
  if (pattern == '') {
    insertInfoAlert('请输入搜索内容');
    search_input.focus();
    return;
  }

  if (searchFor == 'tags') {
    searchTags();
  }
});

function searchTags() {
  // getTags 返回标签集合， addPrefix 把集合数组转化为字符串。
  const tagSet = getTags(search_input);
  const tags = addPrefix(tagSet);
  const url = '/api/search/tags/' + encodeURIComponent(tags);
  loading.text('searching: ' + addPrefix(tagSet, '#'));
  ajaxGet(url, search_btn, that => {
    $('.alert').remove();
    $('#notes-count').text(`找到 ${that.response.length} 篇笔记`);
    refreshNoteList(that.response);
  }, null, function() {
    // not200
    note_list.html('');
  });
}

function refreshNoteList(notes) {
  note_list.html('');
  notes.forEach(note => {
    let updatedAt = dayjs(note.UpdatedAt);
    let item = $('#li-tmpl').contents().clone();
    item.find('.id').text(note.ID);
    item.find('.datetime').text(updatedAt.format('MMM D, HH:mm'));
    item.find('.title')
      .attr('href', '/html/note?id='+note.ID)
      .text(note.Title);
    item.find('.tags').text(addPrefix(note.Tags, '#'));
    item.appendTo(note_list);
  });
}