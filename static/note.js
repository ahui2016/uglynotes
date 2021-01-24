const id = getUrlParam('id');

const edit_btn = $('#edit');
const confirm_block = $('#confirm-block');
const del_confirm_msg = $('#delete-confirm-msg');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');
const undelete_btn = $('#undelete');
const undel_confirm_block = $('#undel-confirm-block');
const undel_block = $('#undelete-block');
const undel_yes_btn = $('#undel-yes');
const undel_no_btn = $('#undel-no');

let isDeleted = false;

let note;
ajaxGet('/api/note/'+id, null, that => {
  note = that.response;
  
  note.Contents = note.Patches.reduce((patched, patch) => {
    return patched = Diff.applyPatch(patched, patch)}, "");

  const createdAt = dayjs(note.CreatedAt);
  const updatedAt = dayjs(note.UpdatedAt);
  $('#created-at').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#updated-at').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#id').text(note.ID);
  $('#note-type').text(note.Type);
  $('#size').text(fileSizeToString(note.Size));
  edit_btn.attr('href', '/html/note/edit?id='+note.ID);
  $('#history').attr('href', '/html/history?id='+note.ID)

  if (note.Deleted) {
    isDeleted = true;
    edit_btn.removeAttr('href');
    delete_btn.text('Delete forever');
    del_confirm_msg.text('delete this note permanently?');
    undel_block.show();
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

// 恢复按钮
undelete_btn.click(undelete_toggle)
// 取消恢复
undel_no_btn.click(undelete_toggle)

function undelete_toggle(event) {
  event.preventDefault();
  undelete_btn.toggle();
  undel_confirm_block.toggle();
}

// 确认恢复
undel_yes_btn.click(() => {
  const form = new FormData();
  form.append('deleted', false);
  ajaxPut(form, `/api/note/${id}/deleted`, yes_btn, () => {
    isDeleted = false;
    edit_btn.attr('href', '/html/note/edit?id='+id);
    delete_btn.text('Delete');
    del_confirm_msg.text('delete this note?');
    undel_block.hide();
    insertInfoAlert('该笔记已复原，可正常编辑');
  });
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
  let url = `/api/note/${id}/deleted`;
  let msg = `笔记 id:${id} 已删除`;
  if (isDeleted) {
    url = '/api/note/'+id
    msg = `笔记 id:${id} 已彻底删除`;
  }
  function onDelete() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('#title-block').hide();
    $('.contents').hide();
    insertSuccessAlert(msg);
  }
  if (isDeleted) {
    ajaxDelete(url, yes_btn, onDelete);
    return
  }
  const form = new FormData();
  form.append('deleted', true);
  ajaxPut(form, url, yes_btn, onDelete);
});

