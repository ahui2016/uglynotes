const filter = getUrlParam('filter');

let url = '/api/note/all';
let not_found_msg = '数据库中没有笔记';

if (filter == 'deleted') {
  url = '/api/note/deleted';
  not_found_msg = '数据库中没有标记为"已删除"的笔记'
}

ajaxGet(url, null, that => {
  if (!that.response || that.response.length == 0) {
    insertInfoAlert(not_found_msg);
    return;
  }
  that.response.forEach(addNoteElem);
}, () => {
  // onloadend
  $('#loading').hide();
});

ajaxGet("/api/note/all/size", null, that => {
  const totalSize = that.response.TotalSize;
  const capacity = that.response.Capacity;
  const used = fileSizeToString(totalSize, 0);
  const available = fileSizeToString(capacity - totalSize, 0);
  $('#notes-size').text(`已用: ${used}, 剩余可用: ${available}`);
});
