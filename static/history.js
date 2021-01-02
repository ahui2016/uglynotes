let isProtected = false;

const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');

const id = getUrlParam('id');
let note_id;

ajaxGet('/api/history/'+id, null, that => {
  const history = that.response;
  const createdAt = dayjs(history.CreatedAt);
  $('#datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#note-id').text('id:'+history.NoteID);
  $('#history-id').text(history.ID);
  $('#size').text(fileSizeToString(history.Size));
  if (!note_id) {
    note_id = history.NoteID;
    $('#histories').attr('href', '/html/note/history?id='+note_id);
  }

  const protected = $('#protected');
  const protectBtn = $('#protect');
  const unprotectBtn = $('#unprotect');
  protectBtn.click(setProtected);
  unprotectBtn.click(setProtected);
  if (history.Protected) protectToggle();

  function protectToggle() {
    isProtected = !isProtected;
    protected.toggle();
    protectBtn.toggle();
    unprotectBtn.toggle();
  }

  function setProtected(event) {
    const form = new FormData();
    form.append("id", history.ID);
    form.append("protected", !isProtected);
    ajaxPut(
        form, '/api/history/protected', $(event.currentTarget), () => {
      protectToggle();
    });
  }

  const plaintext = $('.plaintext.contents');
  plaintext.text(history.Contents);

  const markdown = $('.markdown.contents');
  const dirty = marked(history.Contents);
  const clean = DOMPurify.sanitize(dirty);
  markdown.html(clean);

  $('input[name="note-type"]').change(() => {
    plaintext.toggle();
    markdown.toggle();
  });
  
}, function() {
  //onloadend
  $('#loading').hide();
});


// 删除按钮
delete_btn.click(event => {
  event.preventDefault();
  delete_btn.hide();
  confirm_block.show();
});

// 取消删除
no_btn.click(event => {
  event.preventDefault();
  confirm_block.hide();
  delete_btn.show();
});

// 确认删除
yes_btn.click(event => {
  event.preventDefault();
  let form = new FormData();
  form.append('id', id);
  ajaxDelete(form, '/api/note/'+id, yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('#title-block').hide();
    $('#display-options').hide();
    $('.contents').hide();
    insertSuccessAlert(`历史版本 id:${id} 已删除`);
  });
});
