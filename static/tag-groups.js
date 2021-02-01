ajaxGet('/api/tag/group/all', null, that => {
  if (!that.response) {
    $('#count').text(0);
    return;
  }
  $('#count').text(that.response.length);

  that.response.forEach(addTagGroup);
}, () => {
  // onloadend
  $('#loading').hide();
  $('#count-block').show();
});

function addTagGroup(group) {
  const item_id = 'item-'+group.ID;
  const tagsString = addPrefix(group.Tags);
  const encodedTags = encodeURIComponent(tagsString);
  const updatedAt = dayjs(group.UpdatedAt);

  const item = $('#li-tmpl').contents().clone();
  item.insertAfter('#li-tmpl');
  const groupElem = item.find('.group');
  const protect = item.find('.protect');
  const unprotect = item.find('.unprotect');
  const protected = item.find('.protected');
  const deleted = item.find('.deleted');
  const delete_btn = item.find('.delete');
  const confirm_block = item.find('.confirm-block');
  const no_btn = item.find('.no-btn');
  const yes_btn = item.find('.yes-btn');
  const tagsElem = item.find('.tags');

  item.attr('id', item_id);
  item.find('.datetime').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
  groupElem.attr('href','/html/search?tags=' + encodedTags);
  group.Tags.forEach(tag => {
    const tagElem = $('#tag-tmpl').contents().clone();
    tagElem
      .text(tag)
      .attr('href', '/html/tag/?name=' + encodeURIComponent(tag));
    tagElem.insertBefore(groupElem);
  });
  tagsElem.text(addPrefix(group.Tags, '#'));

  const toggle_protect = function() {
    protected.toggle();
    protect.toggle();
    unprotect.toggle();
  }
  
  function setProtected(event) {
    const form = new FormData();
    form.append("protected", !group.Protected);
    ajaxPut(form, `/api/tag/group/${group.ID}/protected`, $(event.currentTarget),
      () => {
        toggle_protect();
      }, null, event => {
        // onFail
        insertErrorAlert('操作失败', $('#'+item_id));
      }
    );
  }

  if (group.Protected) toggle_protect();

  protect.click(setProtected);
  unprotect.click(setProtected);

  item.find('.create').attr('href', '/html/note/new?tags=' + encodedTags);
    
  // 删除按钮
  delete_btn.click(delete_toggle);

  // 取消删除
  no_btn.click(delete_toggle);

  function delete_toggle() {
    delete_btn.toggle();
    confirm_block.toggle();
  }

  // 确认删除
  yes_btn.click(event => {
    event.preventDefault();
    ajaxDelete('/api/tag/group/'+group.ID, yes_btn, function() {
      protected.hide();
      deleted.show();
      item.find('.tags').show();
      item.find('.tag').hide();
      item.find('.buttons').hide();
      $('.alert').remove();
    }, null, function() {
      // onFail
      insertErrorAlert('删除失败', $('#'+item_id));
    });
  });

}

// 对标签输入框进行处理
const check_tags_btn = $('#check-tags-btn');
const tags_input = $('#tags-input');
const handle_tags_btn = $('#handle-tags-btn');

tags_input.focus(() => {
  handle_tags_btn.hide();
  check_tags_btn.show();
});
check_tags_btn.click(() => {
  const tags = getTags(tags_input);
  if (tags.size == 0) return ;
  if (tags.size == 1) {
    insertInfoAlert('标签组至少需要 2 个标签');
    return;
  }
  tags_input.val(addPrefix(tags, '#'));
  handle_tags_btn.show();
  check_tags_btn.hide();
});
handle_tags_btn.click(() => {
  const tagsSet = getTags(tags_input);
  if (tagsSet.size < 2) {
    insertInfoAlert('标签组至少需要 2 个标签');
    return;
  }
  const form = new FormData();
  form.append('tags', JSON.stringify(Array.from(tagsSet)));
  ajaxPost(form, '/api/tag/group', handle_tags_btn, that => {
    console.log('group: ', that.response);
    addTagGroup(that.response);
    $('.alert').remove();
    insertSuccessAlert('新标签组添加成功');  
  });
});
