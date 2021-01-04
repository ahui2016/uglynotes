const tagName = $('#name');
const rename = $('#rename');
const name_input = $('#name-input');
const rename_block = $('#rename-block');
const cancel = $('#cancel');
const ok = $('#ok');

const tag_name = getUrlParam('name');
tagName.text(tag_name);

ajaxGet(`/api/tag/${tag_name}/notes`, null, that => {
  if (!that.response) {
    $('#count').text(0);
    return;
  }
  $('#count').text(that.response.length);

  that.response.forEach(note => {
    let updatedAt = dayjs(note.UpdatedAt);
    let item = $('#li-tmpl').contents().clone();
    item.find('.id').text(note.ID);
    item.find('.datetime').text(updatedAt.format('MMM D, HH:mm'));
    item.find('.title')
      .attr('href', '/html/note?id='+note.ID)
      .text(note.Title);
    item.find('.tags').text(addPrefix(note.Tags, '#'));
    item.insertAfter('#li-tmpl');
  });
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
  const new_name = name_input.val().trim();
  if (new_name == '') {
    insertErrorAlert('标签名称不可空白');
    name_input.focus();
    return;
  }

  const form = new FormData();
  const old_name = decodeURIComponent(tag_name);
  form.append('old-name', old_name);
  form.append('new-name', new_name);

  ajaxPut(form, '/api/tag/', ok, () => {
    rename_toggle();
    tagName.val(new_name);
    $('#tag-name').hide();
    $('.alert').hide();
    $('ul').hide();
    insertSuccessAlert(`正在重命名: ${old_name} --> ${new_name}`);
    insertInfoAlert('重命名成功时会自动刷新页面');
    window.setTimeout(function(){
      window.location = '/html/tag/?name=' + encodeURIComponent(new_name)
    }, 3000);
  });  
});
