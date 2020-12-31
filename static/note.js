const id = getUrlParam('id');

ajaxGet('/api/note/'+id, null, that => {
  const note = that.response;
  const updatedAt = dayjs(note.UpdatedAt);
  $('#datetime').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#id').text(note.ID);
  $('#note-type').text(note.Type);
  $('#size').text(fileSizeToString(note.Size));
  $('#edit').attr('href', '/html/note/edit?id='+note.ID);

  if (note.Tags.length > 0) {
    $('#tags').show();
    note.Tags.forEach(tag => {
      const tagElem = $('#tag-tmpl').contents().clone();
      tagElem.text(tag);
      tagElem.insertAfter('#tag-tmpl');    
    });  
  }

  if (note.Type == 'Markdown') {
    const dirty = marked(note.Contents);
    const clean = DOMPurify.sanitize(dirty);
    $('.markdown.contents').show().html(clean);
  } else {
    $('.plaintext.contents').show().text(note.Contents);
  }
}, function() {
  //onloadend
  $('#loading').hide();
});
  