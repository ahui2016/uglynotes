const previewBtn = $('#preview-btn');
const editBtn = $('#edit-btn');
const plaintextBtn = $('#plaintext');
const plaintextLabel = $('#plaintext-label');
const preview = $('#preview');
const textarea = $('#contents');
const tagsElem = $('#tags');
const submit_btn = $('#submit');
const submit_block = $('#submit-block');
const confirm_block = $('#confirm-block');
const update_block = $('#update-block');
const update_btn = $('#update');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');

let note_id = '';
let oldContents = '';
let oldNoteType = '';
let tags = new Set();
let oldTags = new Set();
let autoSubmitID;

$('input[name="note-type"]').change(() => {
  previewBtn.toggle();
});

// 自动在标签前加井号，同时更新全局变量。
$('#tags').blur(() => {
    tags = getTags();
    $('#tags').val(addPrefix(tags, '#'));
});

delete_btn.click(event => {
  event.preventDefault();
  delete_btn.hide();
  confirm_block.show();
});

no_btn.click(event => {
  event.preventDefault();
  confirm_block.hide();
  delete_btn.show();
});

yes_btn.click(event => {
  event.preventDefault();
  let form = new FormData();
  form.append('id', note_id);
  ajaxPost(form, '/note/delete', yes_btn, function() {
    note_id = '';
    $('.alert').hide();
    $('form').hide();
    window.clearInterval(autoSubmitID);
    insertSuccessAlert('笔记已删除');
  });
});

previewBtn.click(event => {
  event.preventDefault();
  const contents = $('#contents').val().trim();
  const dirty = marked(contents);
  const clean = DOMPurify.sanitize(dirty);
  preview.show().html(clean);
  textarea.hide();
  previewBtn.hide();
  editBtn.show();
  plaintextBtn.hide();
  plaintextLabel.hide();
});

editBtn.click(event => {
  textarea.show();
  preview.hide();
  plaintextBtn.show();
  plaintextLabel.show();
  previewBtn.show();
  editBtn.hide();
  textarea.focus();
});

submit_btn.click(submit);
update_btn.click(update);

function submit(event) {
  if (event) event.preventDefault();
  
  let contents = textarea.val().trim();
  if (contents == '') {
    // 有 event 表示点击了按钮，这种情况要给用户提示。
    // 如果没有 event 则是后台自动运行，不需要提示。
    if (event) insertInfoAlert('笔记内容不可空白');
    return;
  }

  let form = new FormData();
  form.append('id', note_id);
  const note_type = $('input[name="note-type"]:checked').val();
  form.append('note-type', note_type);
  form.append('contents', contents);
  form.append('tags', JSON.stringify(Array.from(tags)));

  ajaxPost(form, '/note/new', submit_btn, function(that) {
    note_id = that.response.id;
    oldNoteType = note_type;
    oldContents = contents;
    oldTags = tags;
    submit_block.hide();
    update_block.show();
    insertSuccessAlert('新笔记创建成功');
  });
}

function update(event) {
  if (event) event.preventDefault();
  
  const note_type = $('input[name="note-type"]:checked').val();
  if (note_type != oldNoteType) {
    let form = new FormData();
    form.append('id', note_id);
    form.append('note-type', note_type)
    ajaxPost(form, '/note/type/update', update_btn, function() {
      oldNoteType = note_type;
      insertSuccessAlert('笔记类型更新成功: ' + note_type);
    });
  }

  if (!areSetsEqual(tags, oldTags)) {
    let form = new FormData();
    form.append('id', note_id);
    form.append('tags', JSON.stringify(Array.from(tags)));
    ajaxPost(form, '/note/tags/update', update_btn, function() {
      oldTags = tags;
      insertSuccessAlert('标签更新成功: ' + addPrefix(tags, ''));
    });
  }

  let contents = textarea.val().trim();
  if (contents == '') {
    if (event) insertInfoAlert('笔记内容不可空白');
    return;
  }

  if (contents == oldContents) {
    if (event) insertInfoAlert('笔记内容没有变化');
    return;
  }

  if (contents != oldContents) {
    let form = new FormData();
    form.append('id', note_id);
    form.append('contents', contents);
    ajaxPost(form, '/note/contents/update', update_btn, function(that) {
      oldContents = contents;
      insertHistoryAlert(that.response.id);
    });
  }
}

function submitOrUpdate() {
  if (!note_id) {
    submit();
  } else {
    update();
  }
}

autoSubmitID = window.setInterval(submitOrUpdate, delayOfAutoUpdate);
