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
const copy_btn = $('#copy');

let id = '';
let oldContents = '';
let oldNoteType = 'plaintext';
let tags = new Set();
let oldTags = new Set();
let autoSubmitID;
let autoUpdateCount = 1;

const higher = $('#higher');
higher.click(function(){
  const rows = textarea.attr('rows');
  textarea.attr('rows', parseInt(rows) + 10);
});

const wider = $('#wider');
wider.click(function(){
  const cols = textarea.attr('cols');
  textarea.attr('cols', parseInt(cols) + 10);
});

function toggle_higher_wider() {
  higher.toggle();
  wider.toggle();
}

/* 初始化 note/new 表单 */

if (document.location.pathname == "/html/note/new") {
  loading.hide();
  const param_tags = getUrlParam('tags');
  if (param_tags) {
    tagsElem.val(param_tags);
    tags = getTags();
    tagsElem.val(addPrefix(tags, '#'));
  }
}

/* 初始化 note/edit 表单 */

if (document.location.pathname == "/html/note/edit") {
  const param_id = getUrlParam('id');
  if (param_id) initEditForm(param_id);
}


function initEditForm(param_id) {
  ajaxGet('/api/note/'+param_id, null, that => {
    const note = that.response;

    if (note.Deleted) {
      onFail();
      insertErrorAlert(`the note(id:${param_id}) is deleted`);
      return;
    }

    id = note.ID;
  
    if (note.Type == 'Markdown') {
      plaintextBtn.prop('checked', false);
      $('#markdown').prop('checked', true);
      previewBtn.show();
      oldNoteType = 'markdown';
    }

    note.Contents = note.Patches.reduce((patched, patch) => {
      return patched = Diff.applyPatch(patched, patch)}, "");
  
    textarea.val(note.Contents);
    oldContents = note.Contents;
  
    tagsElem.val(addPrefix(note.Tags, '#'));
    tags = new Set(note.Tags);
    oldTags = new Set(note.Tags);

    updateSize(note.Size);
    enterEditMode(note.Patches.length);
    insertSuccessAlert('已获取笔记 id:' + note.ID.toUpperCase());
    insertSuccessAlert('已进入编辑模式');

  }, onLoadend, onFail);
}

function onLoadend() {
  loading.hide();
}

function onFail() {
  $('form').hide();
  window.clearInterval(autoSubmitID);
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
  toggle_higher_wider();
  copy_btn.hide();
});

// 编辑按钮
editBtn.click(event => {
  textarea.show();
  preview.hide();
  plaintextBtn.show();
  plaintextLabel.show();
  previewBtn.show();
  editBtn.hide();
  toggle_higher_wider();
  copy_btn.show();
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
  if (contents.length > NoteSizeLimit) {
    insertErrorAlert('超过笔记体积上限: ' + fileSizeToString(NoteSizeLimit));
    return;
  }
  if (tags.size < 2) {
    insertErrorAlert('至少需要 2 个标签才能创建笔记');
    return;
  }

  const patch = Diff.createPatch(" ", "", contents);
  const form = new FormData();
  const note_type = $('input[name="note-type"]:checked').val();
  form.append('note-type', note_type);
  form.append('title', contents.substring(0, NoteTitleLimit));
  form.append('patch', patch);
  form.append('tags', JSON.stringify(Array.from(tags)));

  if (!event) autoUpdateCount++;
  ajaxPost(form, '/api/note', submit_btn, function(that) {
    id = that.response.message;
    oldNoteType = note_type;
    oldContents = contents;
    oldTags = tags;
    updateSize(contents.length);
    enterEditMode(1);
    insertSuccessAlert('新笔记创建成功 id:' + id);
  });
}

function enterEditMode(version) {
  $('title').text('Edit Note .. uglynotes');
  $('#where').text('Edit Note');
  $('#head-buttons').show();
  $('#readonly-mode').attr('href', '/html/note?id='+id);
  $('#history').attr('href', `/html/history?id=${id}&version=${version}`);
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
    form.append('note-type', note_type)

    if (!event) autoUpdateCount++;
    ajaxPut(form, `/api/note/${id}/type`, update_btn, function() {
      oldNoteType = note_type;
      insertSuccessAlert('笔记类型更新成功: ' + note_type);
    });
  }

  // 更新标签
  if (!setsAreEqual(tags, oldTags)) {
    if (tags.size < 2) {
      insertErrorAlert('至少需要 2 个标签');
      return;
    }
    const form = new FormData();
    form.append('tags', JSON.stringify(Array.from(tags)));

    if (!event) autoUpdateCount++;
    ajaxPut(form, `/api/note/${id}/tags`, update_btn, function() {
      oldTags = tags;
      insertSuccessAlert('标签更新成功: ' + addPrefix(tags, ''));
    });
  }

  let contents = textarea.val().trim();
  if (contents == '') {
    if (event) insertInfoAlert('笔记内容不可空白');
    return;
  }
  if (contents.length > NoteSizeLimit) {
    insertErrorAlert('超过笔记体积上限: ' + fileSizeToString(NoteSizeLimit, 0));
    return;
  }
  if (contents == oldContents) {
    if (event) insertInfoAlert('笔记内容没有变化');
    return;
  }

  // 更新笔记内容
  if (contents != oldContents) {
    const patch = Diff.createPatch(" ", oldContents, contents);
    const form = new FormData();
    form.append('title', contents.substring(0, NoteTitleLimit));
    form.append('patch', patch);

    if (!event) autoUpdateCount++;
    ajaxPatch(form, '/api/note/'+id, update_btn, that => {
      oldContents = contents;
      updateSize(contents.length);
      let count = that.response.message;
      $('#history').attr('href', `/html/history?id=${id}&version=${count}`);
      insertSuccessAlert(`笔记内容更新，产生第 ${count} 个历史版本`);
    });
  }
}

// 复制至剪贴板
const clipboard = new ClipboardJS('#copy');
clipboard.on('success', e => {
  insertSuccessAlert('笔记内容已复制到剪贴板', loading);
});
clipboard.on('error', e => {
  console.error('Action:', e.action);
  console.error('Trigger:', e.trigger);
  insertErrorAlert('复制失败，详细信息见控制台', loading);
});

// 自动更新
function submitOrUpdate() {
  if (!id) {
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

function updateSize(size) {
  $('#note-size').text('Size: ' + fileSizeToString(size));
} 

autoSubmitID = window.setInterval(submitOrUpdate, DelayOfAutoUpdate);

// 用于初始化表单。
function initAjaxGet(url, onload, onloadend) {
  let xhr = new XMLHttpRequest();
  xhr.responseType = 'json';
  xhr.open('GET', url);
  xhr.onerror = function () {
    window.alert('An error occurred during the transaction');
  };
  xhr.onload = onload;
  xhr.addEventListener('loadend', function() {
    if (onloadend) onloadend(this);
  });
  xhr.send();
}
