const id = getUrlParam('id');

$('#note-id')
  .text('id:'+id)
  .attr('href', '/html/note?id='+id);

let history_size = 0;

ajaxGet(`/api/note/${id}/history`, null, that => {
  if (!that.response) {
    insertInfoAlert('该笔记没有历史版本');
    $('#size-block').hide();
    $('#head-buttons').hide();
    return;
  }
  that.response.forEach(history => {
    history_size += history.Size;

    const item = $('#li-tmpl').contents().clone();
    item.insertAfter('#li-tmpl');

    item.find('.id').text(history.ID);
    if (history.Protected) {
      item.find('.protected').show();
    }
    const createdAt = dayjs(history.CreatedAt);
    item.find('.datetime').text(createdAt.format('MMM D, HH:mm'));
    item.find('.size').text(fileSizeToString(history.Size));
    item.find('.title')
      .text(history.Contents)
      .attr('href', `/html/history?id=${history.ID}&noteid=${id}`);
  });

  $('#history-size')
    .show()
    .text(fileSizeToString(history_size));
}, () => {
  // onloadend
  $('#loading').hide();
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
  ajaxDelete(`/api/note/${id}/history`, yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('p').hide();
    $('ul').hide();
    insertSuccessAlert(`正在删除未保护历史...`);
    insertInfoAlert('删除成功时会自动刷新页面');
    window.setTimeout(function(){
      window.location.reload();
    }, 5000);
  });
});
