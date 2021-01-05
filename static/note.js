const id = getUrlParam('id');

const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');

ajaxGet('/api/note/'+id, null, that => {
  const note = that.response;
  const updatedAt = dayjs(note.UpdatedAt);
  $('#datetime').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#id').text(note.ID);
  $('#note-type').text(note.Type);
  $('#size').text(fileSizeToString(note.Size));
  $('#edit').attr('href', '/html/note/edit?id='+note.ID);
  $('#history').attr('href', '/html/note/history?id='+note.ID)

  if (note.Tags && note.Tags.length > 0) {
    $('#tags').show();
    note.Tags.forEach(tag => {
      const tagElem = $('#tag-tmpl').contents().clone();
      tagElem
        .text(tag)
        .attr('href', '/html/tag/?name=' + encodeURIComponent(tag));
      tagElem.insertBefore('#tag-tmpl');    
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


// 删除按钮
delete_btn.click(delete_toggle);

// 取消删除
no_btn.click(delete_toggle);

function delete_toggle(event) {
  event.preventDefault();
  delete_btn.toggle();
  confirm_block.toggle();
}

// 确认删除
yes_btn.click(event => {
  event.preventDefault();
  ajaxDelete('/api/note/'+id, yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('#title-block').hide();
    $('.contents').hide();
    insertSuccessAlert(`笔记 id:${id} 已删除`);
  });
});
