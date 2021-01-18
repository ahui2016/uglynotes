const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');
const diff = $('.diff');
const number_input = $('#number');
const go_btn = $('#go-btn');
const export_btn = $('#export-btn');
const previous_btn = $('#previous-btn');
const next_btn = $('#next-btn');

const id = getUrlParam('id');
let note, current_n, max_n;

ajaxGet('/api/note2/'+id, null, that => {
  note = that.response;
  $('#note-id')
    .text('id:'+id)
    .attr('href', '/html/note?id='+id);
  max_n = note.Patches.length;
  number_input.val(max_n).attr('max', max_n);
  gotoHistory(max_n);
}, function() {
  //onloadend
  $('#loading').hide();
});

function gotoHistory(n) {
  if (current_n == n) return;
  current_n = n;
  const diffString = note.Patches[n-1];
  const diffJson = Diff2Html.parse(diffString);
  const diffHtml = Diff2Html.html(diffJson, { 
    drawFileList: false,
  });
  diff.html(diffHtml);
}

go_btn.click(() => {
  const n = number_input.val();
  gotoHistory(current_n);
});

previous_btn.click(() => {
  if (current_n == 1) return;
  const n = current_n - 1;
  number_input.val(n);
  gotoHistory(n);
});

next_btn.click(() => {
  if (current_n == max_n) return;
  const n = current_n + 1;
  number_input.val(n);
  gotoHistory(n);
});

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
  let form = new FormData();
  form.append('id', id);
  ajaxDelete('/api/history/'+id, yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('#title-block').hide();
    $('#display-options').hide();
    $('.contents').hide();
    insertSuccessAlert(`历史版本 id:${id} 已删除`);
  });
});
