const tagGroup = getUrlParam('tags');

const search_input = $('#search-input');
const search_btn = $('#search-btn');
const loading = $('#loading');
const note_list = $('ul');
const notesCount = $('#notes-count');

search_btn.click(event => {
  event.preventDefault();
  const pattern = search_input.val().trim();
  if (pattern == '') {
    insertInfoAlert('请输入搜索内容');
    search_input.focus();
    return;
  }

  const searchBy = $('input[name="search-by"]:checked').val()
  if (searchBy == 'tags') searchTags();
  if (searchBy == 'title') searchTitle();
});

function searchTags() {
  // getTags 返回标签集合， addPrefix 把集合数组转化为字符串。
  const tagSet = getTags(search_input);
  const tags = addPrefix(tagSet);
  const url = '/api/search/tags/' + encodeURIComponent(tags);
  loading.text('searching: ' + addPrefix(tagSet, '#'));
  ajaxGet(url, search_btn, onSuccess, null, onFail);
}

function searchTitle() {
  const title = search_input.val().trim();
  const url = '/api/search/title/' + encodeURIComponent(title);
  loading.text('searching: ' + title);
  ajaxGet(url, search_btn, onSuccess, null, onFail);
}

function onSuccess(that) {
  const notes = getNotes(that);
  refreshNoteList(notes);
}

function onFail(that) {
  note_list.html('');
  notesCount.hide();
}

function getNotes(that) {
  $('.alert').remove();
  let notes = [];
  if (that.response) notes = that.response;
  notesCount
    .show()
    .text(`找到 ${notes.length} 篇笔记`);
  return notes;
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

// 当网址中带有标签组参数时，直接自动搜索
if (tagGroup) {
  search_input.val(tagGroup);
  search_btn.click();
}
