const id = getUrlParam('id');

const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');

let isDeleted = false;

ajaxGet('/api/note/'+id, null, that => {
  const note = that.response;
  
  note.Contents = note.Patches.reduce((patched, patch) => {
    return patched = Diff.applyPatch(patched, patch)}, "");

  const updatedAt = dayjs(note.UpdatedAt);
  $('#datetime').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#id').text(note.ID);
  $('#note-type').text(note.Type);
  $('#size').text(fileSizeToString(note.Size));
  if (!note.Deleted) $('#edit').attr('href', '/html/note/edit?id='+note.ID);
  $('#history').attr('href', '/html/history?id='+note.ID)

  if (note.Deleted) {
    isDeleted = true;
    delete_btn.text('Delete forever');
    insertInfoAlert('该笔记被标记为 “已删除”');
  }

  if (note.Tags && note.Tags.length > 0) {
    $('#tags').show();
    note.Tags.forEach(tag => {
      const tagElem = $('#tag-tmpl').contents().clone();
      tagElem
        .text(tag)
        .attr('href', '/html/tag/?name=' + encodeURIComponent(tag));
      tagElem.insertBefore('#tag-tmpl');
    });
    if (note.Tags.length > 1) {
      const tagElem = $('#tag-tmpl').contents().clone();
      tagElem
        .text('group')
        .addClass('group')
        .attr('title', 'search tag group')
        .attr('href', '/html/search?tags=' + encodeURIComponent(addPrefix(note.Tags)));
      tagElem.insertBefore('#tag-tmpl');    
    }
  }

  const clipboard = new ClipboardJS('#copy', {
    text: () => { return note.Contents; }
  });
  clipboard.on('success', () => {
    insertSuccessAlert('笔记内容已复制到剪贴板');
  });
  clipboard.on('error', e => {
    console.error('Action:', e.action);
    console.error('Trigger:', e.trigger);
    insertErrorAlert('复制失败，详细信息见控制台');
  });

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
  let url = '/api/note/'+id;
  let msg = `笔记 id:${id} 已删除`;
  if (isDeleted) {
    url = '/api/note/deleted/'+id
    msg = `笔记 id:${id} 已彻底删除`;
  }
  ajaxDelete(url, yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('#title-block').hide();
    $('.contents').hide();
    insertSuccessAlert(msg);
  });
});
