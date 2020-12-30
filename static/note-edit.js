const loading = $('#loading');
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
let oldNoteType = 'plaintext';
let tags = new Set();
let oldTags = new Set();
let autoSubmitID;
let autoUpdateCount = 1;

/* 开始初始化表单 */

const param_id = getUrlParam('id');

if (!param_id) loading.hide();

if (param_id) {
  const form = new FormData();
  form.append('id', param_id);

  // init 函数定义在本文件末尾。
  init(form, '/api/note', function() {
    if (this.status == 200) {
      const note = this.response;
      note_id = note.ID;
    
      if (note.Type == 'Markdown') {
        plaintextBtn.prop('checked', false);
        $('#markdown').prop('checked', true);
        previewBtn.show();
        oldNoteType = 'markdown';
      }
    
      textarea.val(note.Contents);
      oldContents = note.Contents;
    
      tagsElem.val(addPrefix(note.Tags, '#'));
      tags = new Set(note.Tags);
      oldTags = new Set(note.Tags);

      enterEditMode();
      insertSuccessAlert('已获取笔记 id:' + note.ID.toUpperCase());
      insertSuccessAlert('已进入编辑模式');
    } else {
      $('.alert').hide();
      $('form').hide();
      window.clearInterval(autoSubmitID);
      let errMsg = !this.response ? this.status : this.response.message;
      errMsg = `[id:${param_id}] ` + errMsg
      insertErrorAlert(errMsg);
    }
  }, function() {
    // onloadend
    loading.hide();
  });
}

/* 表单初始化结束 */

$('input[name="note-type"]').change(() => {
  previewBtn.toggle();
});

// 自动在标签前加井号，同时更新全局变量。
tagsElem.blur(() => {
    tags = getTags();
    tagsElem.val(addPrefix(tags, '#'));
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
  form.append('id', note_id);
  ajaxPost(form, '/api/note/delete', yes_btn, function() {
    note_id = '';
    $('.alert').hide();
    $('form').hide();
    window.clearInterval(autoSubmitID);
    insertSuccessAlert('笔记已删除');
  });
});

// 预览按钮
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

// 编辑按钮
editBtn.click(event => {
  textarea.show();
  preview.hide();
  plaintextBtn.show();
  plaintextLabel.show();
  previewBtn.show();
  editBtn.hide();
  textarea.focus();
});

// 提交按钮和更新按钮
submit_btn.click(submit);
update_btn.click(update);

// 创建新笔记
function submit(event) {
  if (event) event.preventDefault();
  
  let contents = textarea.val().trim();
  if (contents == '') {
    // 有 event 表示点击了按钮，这种情况要给用户提示。
    // 如果没有 event 则是后台自动运行，不需要提示。
    if (event) insertInfoAlert('笔记内容不可空白');
    return;
  }

  const form = new FormData();
  const note_type = $('input[name="note-type"]:checked').val();
  form.append('note-type', note_type);
  form.append('contents', contents);
  form.append('tags', JSON.stringify(Array.from(tags)));

  if (!event) autoUpdateCount++;
  ajaxPost(form, '/api/note/new', submit_btn, function(that) {
    note_id = that.response;
    oldNoteType = note_type;
    oldContents = contents;
    oldTags = tags;
    enterEditMode();
    insertSuccessAlert('新笔记创建成功');
  });
}

function enterEditMode() {
  $('#where').text('Edit Note');
  $('#readonly-mode')
    .show()
    .attr('href', '/html/note?id='+note_id);
  submit_block.hide();
  update_block.show();
}

// 更新
function update(event) {
  if (event) event.preventDefault();
  
  // 更新笔记类型
  const note_type = $('input[name="note-type"]:checked').val();
  if (note_type != oldNoteType) {
    const form = new FormData();
    form.append('id', note_id);
    form.append('note-type', note_type)

    if (!event) autoUpdateCount++;
    ajaxPost(form, '/api/note/type/update', update_btn, function() {
      oldNoteType = note_type;
      insertSuccessAlert('笔记类型更新成功: ' + note_type);
    });
  }

  // 更新标签
  if (!setsAreEqual(tags, oldTags)) {
    const form = new FormData();
    form.append('id', note_id);
    form.append('tags', JSON.stringify(Array.from(tags)));

    if (!event) autoUpdateCount++;
    ajaxPost(form, '/api/note/tags/update', update_btn, function() {
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

  // 更新笔记内容
  if (contents != oldContents) {
    const form = new FormData();
    form.append('id', note_id);
    form.append('contents', contents);

    if (!event) autoUpdateCount++;
    ajaxPost(form, '/api/note/contents/update', update_btn, function(that) {
      oldContents = contents;
      insertHistoryAlert(that.response.id);
    });
  }
}

// 自动更新
function submitOrUpdate() {
  if (!note_id) {
    submit();
  } else {
    update();
  }
  if (autoUpdateCount > AutoUpdateLimit) {
    insertErrorAlert('已达到自动更新次数上限，请手动更新或刷新页面。');
    window.clearInterval(autoSubmitID);
    return;
  }
}

autoSubmitID = window.setInterval(submitOrUpdate, DelayOfAutoUpdate);


// 初始化表单。
function init(form, url, onload, onloadend) {
  let xhr = new XMLHttpRequest();
  xhr.responseType = 'json';
  xhr.open('POST', url);
  xhr.onerror = function () {
    window.alert('An error occurred during the transaction');
  };
  xhr.onload = onload;
  xhr.addEventListener('loadend', function() {
    if (onloadend) onloadend(this);
  });
  xhr.send(form);
}
