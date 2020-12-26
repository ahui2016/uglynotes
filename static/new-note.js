const previewBtn = $('#preview-btn');
const editBtn = $('#edit-btn');
const plaintextBtn = $('#plaintext');
const plaintextLabel = $('#plaintext-label');
const preview = $('#preview');
const textarea = $('#contents');
const tagsElem = $('#tags');
const submit_btn = $('#submit');
const submit_block = $('#submit-block');
const update_block = $('#update-block');
const update_btn = $('#update');
const delete_btn = $('#delete');

let note_id = '';
let tags;
let oldContents = '';
let oldTags = '';
let oldNoteType = '';

$('input[name="note-type"]').change(() => {
  previewBtn.toggle();
});

// 自动在标签前加井号，同时更新全局变量。
$('#tags').blur(() => {
    tags = getTags();
    $('#tags').val(addPrefix(tags, '#'));
});

previewBtn.click(event => {
  event.preventDefault();
  const markdown = $('#contents').val();
  const dirty = marked(markdown);
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
  form.append('tags', getTags());

  ajaxPost(form, '/note/new', submit_btn, function(that) {
    note_id = that.response.id;
    oldNoteType = note_type;
    oldContents = contents;
    oldTags = tagsElem.val();
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
      insertInfoAlert('笔记类型更新成功: ' + note_type);
    });
  }

  const tagsString = tagsElem.val();
  if (tagsString != oldTags) {
    let form = new FormData();
    form.append('id', note_id);
    form.append('tags', getTags());
    ajaxPost(form, '/note/tags/update', update_btn, function() {
      oldTags = tagsString;
      insertInfoAlert('笔记的标签更新成功');
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

window.setInterval(submitOrUpdate, delayOfAutoUpdate);

