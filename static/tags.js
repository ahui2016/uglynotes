ajaxGet('/api/tag/all', null, that => {
  if (!that.response) {
    $('#count').text(0);
    return;
  }
  $('#count').text(that.response.length);

  that.response.forEach(tag => {
    addTagItem(tag, '#sort-by-name>div');
  });
}, () => {
  // onloadend
  $('#loading').hide();
  $('#sort-by').show();
  $('#count-block').show();
});

ajaxGet('/api/tag/all-by-date', null, that => {
  that.response.forEach(tag => {
    addTagItem(tag, '#sort-by-date>div');
  });
});

function addTagItem(tag, insertPoint) {
  const createdAt = dayjs(tag.CreatedAt);
  const item = $('#li-tmpl').contents().clone();

  if (insertPoint == '#sort-by-date>div') {
    item.insertAfter(insertPoint);
  } else {
    item.insertBefore(insertPoint);
  }

  item.find('.datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
  nameElem = item.find('.name');
  nameElem
    .text(tag.Name)
    .attr('href', '/html/tag/?name=' + encodeURIComponent(tag.Name));
  
  const count = tag.NoteIDs.length;
  item.find('.count').text(count);
  if (count == 0) {
    const delBtnBlock = $('#del-btn-tmpl').contents().clone();
    delBtnBlock.appendTo(item);

    const delete_btn = item.find('.delete');
    const confirm_block = item.find('.confirm-block');
    const no_btn = item.find('.no-btn');
    const yes_btn = item.find('.yes-btn');
    const deleted = item.find('.deleted');
    const del_btn_block = item.find('.del-btn-block');

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
      event.preventDefault();
      ajaxDelete('/api/tag/'+encodeURIComponent(tag.Name), yes_btn, function() {
        $('.alert').hide();
        nameElem.removeAttr('href');
        del_btn_block.hide();
        deleted.show();
      }, null, function() {
        // onFail
        const insertPoint = $(event.currentTarget).parent().parent();
        insertErrorAlert('删除失败', insertPoint);
      });
    });
  }
}

$('#by-date').prop('checked', true).click();

$('input[name="sort-by"]').change(() => {
  $('#sort-by-name').toggle();
  $('#sort-by-date').toggle();
});
