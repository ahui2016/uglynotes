let isProtected = false;

const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');
const plaintext = $('.plaintext.contents');
const markdown = $('.markdown.contents');

const id = getUrlParam('id');
const note_id = getUrlParam('noteid');

ajaxGet('/api/history/'+id, null, that => {
  const history = that.response;
  const createdAt = dayjs(history.CreatedAt);
  $('#datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#note-id').text('id:'+history.NoteID);
  $('#history-id').text(history.ID);
  $('#size').text(fileSizeToString(history.Size));
  $('#histories').attr('href', '/html/note/history?id='+note_id);

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

  plaintext.text(history.Contents);

  const dirty = marked(history.Contents);
  const clean = DOMPurify.sanitize(dirty);
  markdown.html(clean);
  
  ajaxGet('/api/note/'+note_id, null, that => {
    const current_contents = that.response.Contents;
    const diffString = Diff.createPatch(
      " ", current_contents, history.Contents
    );
    const diffJson = Diff2Html.parse(diffString);
    const diffHtml = Diff2Html.html(diffJson, { 
      drawFileList: false 
    });
    $('.diff.contents').html(diffHtml);
  });
}, function() {
  //onloadend
  $('#loading').hide();
});

$('input[name="note-type"]').change(event => {
  const value = event.currentTarget.value;
  plaintext.toggle();
  markdown.toggle();
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
