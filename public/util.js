// 创建历史版本的间隔时间
const DelayOfAutoUpdate = 1000 * 60 * 5 // 5分钟

// 自动保存（自动更新）次数的上限
const AutoUpdateLimit = 100

// NoteSizeLimit 限制每篇笔记的体积。
// 注意：该限制还需要在 settings.go 中设置（为了做后端限制）
const NoteSizeLimit = 1 << 19 // 512 KB

// 插入出错提示
function insertErrorAlert(msg, where) {
  insertAlert('danger', msg, where);
}

// 插入普通提示
function insertInfoAlert(msg, where) {
  insertAlert('info', msg, where);
}

// 插入成功提示
function insertSuccessAlert(msg, where) {
  insertAlert('success', msg, where);
}

// 插入提示
function insertAlert(type, msg, where) {
  console.log(msg);
  let alertElem = $('#alert-'+type+'-tmpl').contents().clone();
  alertElem.find('.alert-time').text(dayjs().format('HH:mm:ss'));
  alertElem.find('.alert-message').text(msg);
  if (!where) where = '#alert-insert-after-here';
  alertElem.find('.alert-dismiss').click(event => {
    $(event.currentTarget).parent().remove();
  });
  alertElem.insertAfter(where);
}

// 向服务器提交表单，在等待过程中 btn 会失效，避免重复提交。
function ajaxDo(method, form, url, btn, onSuccess, onloadend, onFail) {
  if (btn) {
    btn.prop('disabled', true);
  }
  let xhr = new XMLHttpRequest();
  xhr.responseType = 'json';
  xhr.open(method, url);
  xhr.onerror = function () {
    window.alert('An error occurred during the transaction');
  };
  xhr.addEventListener('load', function() {
    if (this.status == 200) {
      if (onSuccess) onSuccess(this);
    } else {
        let errMsg = !this.response ? this.status : this.response.message;
        insertErrorAlert(errMsg);
        if (onFail) onFail(this);
    }
  });
  xhr.addEventListener('loadend', function() {
    if (btn) {
      btn.prop('disabled', false);
    }
    if (onloadend) onloadend(this);
  });
  
  if (method.toUpperCase() == 'GET') {
    xhr.send();
    return;
  }
  xhr.send(form);
}

function ajaxPost(form, url, btn, onSuccess, onloadend) {
  ajaxDo('POST', form, url, btn, onSuccess, onloadend);
}

function ajaxPut(form, url, btn, onSuccess, onloadend) {
  ajaxDo('PUT', form, url, btn, onSuccess, onloadend);
}

function ajaxDelete(url, btn, onSuccess, onloadend, onFail) {
  ajaxDo('DELETE', null, url, btn, onSuccess, onloadend, onFail);
}

function ajaxGet(url, btn, onSuccess, onloadend, onFail) {
  ajaxDo('GET', null, url, btn, onSuccess, onloadend, onFail);
}

// 获取地址栏的参数。
function getUrlParam(param) {
  let loc = new URL(document.location);
  return loc.searchParams.get(param);
}

// 把标签文本框内的字符串转化为集合。
function getTags(tagsElem) {
  if (!tagsElem) {
    tagsElem = $('#tags');
  }
  let trimmed = tag_replace(tagsElem.val());
  if (trimmed.length == 0) {
    return new Set();
  }
  let arr = trimmed.split(/ +/);
  return new Set(arr);
}

function tag_replace(tags) {
  return tags.replace(/[#;,，'"/\+\n]/g, ' ').trim();
}

// 把集合数组转化为字符串。
function addPrefix(aSet, prefix) {
  if (!aSet) return '';
  let arr = Array.from(aSet);
  if (!prefix) prefix = '';
  return arr.map(x => prefix + x).join(' ');
}

function setsAreEqual(a, b) {
  if (a.size !== b.size) return false;
  for (const item of a) if (!b.has(item)) return false;
  return true;
}

// 把文件大小换算为 KB 或 MB
function fileSizeToString(fileSize, fixed) {
  if (fixed == null) {
    fixed = 2
  }
  const sizeMB = fileSize / 1024 / 1024;
  if (sizeMB < 1) {
      return `${(sizeMB * 1024).toFixed(fixed)} KB`;
  }
  return `${sizeMB.toFixed(fixed)} MB`;
}

function addNoteElem(note) {
  let updatedAt = dayjs(note.UpdatedAt);
  let item = $('#li-tmpl').contents().clone();
  item.find('.id').text(note.ID);
  item.find('.datetime').text(updatedAt.format('MMM D, HH:mm'));
  const titleElem = item.find('.title');
  titleElem
    .attr('href', '/html/note?id='+note.ID)
    .text(note.Title);
  item.find('.tags').text(addPrefix(note.Tags, '#'));
  item.prependTo('ul');

  const deleted = item.find('.deleted');
  const del_btn_block = item.find('.del-btn-block');
  const delete_btn = item.find('.delete');
  const confirm_block = item.find('.confirm-block');
  const no_btn = item.find('.no-btn');
  const yes_btn = item.find('.yes-btn');

  function delete_toggle() {
    delete_btn.toggle();
    confirm_block.toggle();
  }

  // 删除按钮
  delete_btn.click(delete_toggle);

  // 取消删除
  no_btn.click(delete_toggle);

  // 确认删除
  yes_btn.click(event => {
    ajaxDelete('/api/note/'+note.ID, yes_btn, function() {
      $('.alert').hide();
      titleElem.removeAttr('href');
      del_btn_block.hide();
      deleted.show();
    }, null, function() {
      // onFail
      const insertPoint = $(event.currentTarget).parent().parent().parent();
      insertErrorAlert('删除失败', insertPoint);
    });
  });
}

// 初始化页面说明。
$('#about-page-icon').click(() => {
  $('#about-page-info').toggle();
});
