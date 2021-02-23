const tagName = $('#name');
const tagNameBlock = $('#tag-name-block');
const rename = $('#rename');
const name_input = $('#name-input');
const rename_block = $('#rename-block');
const cancel = $('#cancel');
const ok = $('#ok');
const check_tags_btn = $('#check-tags-btn');

const tag_id = getUrlParam('id');
let tag;
ajaxGet(`/api/tag/${tag_id}`, null, that => {
  tag = that.response;
  tagName.text(tag.Name);
});

ajaxGet(`/api/tag/${tag_id}/notes`, null, that => {
  if (!that.response) {
    insertInfoAlert('找不到相关笔记');
    return;
  }
  tagNameBlock.show();
  $('#count-block').show();
  const notes = that.response;
  if (!notes) {
    $('#count').text(0);
    return;
  }
  $('#count').text(notes.length);

  notes.sort((a, b) => {
    if (a.UpdatedAt > b.UpdatedAt) return 1;
    if (a.UpdatedAt < b.UpdatedAt) return -1;
    return 0;
  });
  notes.forEach(addNoteElem);
}, () => {
  // onloadend
  $('#loading').hide();
});

rename.click(event => {
  event.preventDefault();
  rename_toggle();
  name_input.focus();
});
cancel.click(event => {
  event.preventDefault();
  rename_toggle();
});

function rename_toggle() {
  rename.toggle();
  rename_block.toggle();
}

ok.click(() => {
  const new_name = getTag(name_input);
  if (new_name == '') {
    insertErrorAlert('标签名称不可空白');
    name_input.focus();
    return;
  }

  const form = new FormData();
  form.append('new-name', new_name);

  ajaxPut(form, '/api/tag/' + tag_id, ok, (that) => {
    rename_toggle();
    tagName.text(new_name);
    tagNameBlock.hide();
    $('.alert').hide();
    $('ul').hide();
    insertInfoAlert('重命名成功时会自动刷新页面');
    insertSuccessAlert(`正在重命名: ${tag.Name} --> ${new_name}`);

    const newTagID = that.response.message;
    if (newTagID == tag_id) {
      window.setTimeout(function(){window.location.reload()}, 5000);
    } else {
      window.setTimeout(() => {window.location = '/html/tag?id='+newTagID}, 5000);
    }
  });  
});

const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');

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
  ajaxDelete('/api/tag/'+tag_id, yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('p').hide();
    $('ul').hide();
    insertSuccessAlert(`已删除标签: ${tag.Name}`);
  });
});

// 对标签文本框内的字符串进行处理，提取出一个标签。
function getTag(tagsElem) {
  let trimmed = tag_replace(tagsElem.val());
  if (trimmed.length == 0) {
    return '';
  }
  let arr = trimmed.split(/ +/);
  return arr[0];
}

name_input.focus(() => {
  ok.hide();
  check_tags_btn.show();
});
name_input.blur(() => {
  const tag = getTag(name_input);
  if (tag) {
    name_input.val('#' + tag);
    ok.show();  
    check_tags_btn.hide();
  }
});
