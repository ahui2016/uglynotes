// 创建历史版本的间隔时间
const DelayOfAutoUpdate = 1000 * 10

// setInterval 自动运行次数的上限
const AutoUpdateLimit = 3

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
function ajaxDo(method, form, url, btn, onload, onloadend) {
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
      if (onload) onload(this);
    } else {
        let errMsg = !this.response ? this.status : this.response.message;
        insertErrorAlert(errMsg);
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

function ajaxPost(form, url, btn, onload, onloadend) {
  ajaxDo('POST', form, url, btn, onload, onloadend);
}

function ajaxPut(form, url, btn, onload, onloadend) {
  ajaxDo('PUT', form, url, btn, onload, onloadend);
}

function ajaxDelete(url, btn, onload, onloadend) {
  ajaxDo('DELETE', null, url, btn, onload, onloadend);
}

function ajaxGet(url, btn, onload, onloadend) {
  ajaxDo('GET', null, url, btn, onload, onloadend);
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
  let trimmed = tagsElem.val().replace(/[#;,，\n]/g, ' ').trim();
  if (trimmed.length == 0) {
    return [];
  }
  let arr = trimmed.split(/ +/);
  return new Set(arr);
}

// 把集合数组转化为字符串。
function addPrefix(aSet, prefix) {
  if (!aSet) return '';
  let arr = Array.from(aSet);
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
