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
});

function addTagGroup(group) {
  const tagsString = addPrefix(group.Tags);
  const encodedTags = encodeURIComponent(tagsString);
  const updatedAt = dayjs(group.UpdatedAt);

  const item = $('#li-tmpl').contents().clone();
  const groupElem = item.find('.group');
  const protect = item.find('.protect');
  const unprotect = item.find('.unprotect');
  const protected = item.find('.protected');
  item.insertAfter('#li-tmpl');

  item.find('.datetime').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
  groupElem.attr('href','/html/search?tags=' + encodedTags);
  group.Tags.forEach(tag => {
    const tagElem = $('#tag-tmpl').contents().clone();
    tagElem
      .text(tag)
      .attr('href', '/html/tag/?name=' + encodeURIComponent(tag));
    tagElem.insertBefore(groupElem);
  });

  const toggle_protect = function() {
    protected.toggle();
    protect.toggleClass('enabled');
    unprotect.toggleClass('enabled');
  }
  
  function setProtected(event) {
    const form = new FormData();
    form.append("id", group.ID);
    form.append("protected", !group.Protected);
    ajaxPut(
      form, '/api/tag/group/protected', $(event.currentTarget), () => {
      toggle_protect();
    });
  }

  if (group.Protected) toggle_protect();

  protect.click(setProtected);
  unprotect.click(setProtected);

  item.find('.create').click(() => {
    window.location = '/html/note/new?tags=' + encodedTags;
  });
}

const tags_input = $('#tags-input');
const add_btn = $('#add-btn');

add_btn.click(event => {
  event.preventDefault();
  const tagsSet = getTags(tags_input);
  if (tagsSet.size < 2) {
    insertInfoAlert('标签组至少需要 2 个标签');
    return;
  }
  const form = new FormData();
  form.append('tags', JSON.stringify(Array.from(tagsSet)));
  ajaxPost(form, '/api/tag/group', add_btn, that => {
    addTagGroup(that.response);
    $('.alert').remove();
    insertSuccessAlert('新标签组添加成功');  
  });
});