ajaxGet('/api/note/all', null, that => {
  that.response.forEach(addNoteElem);
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
