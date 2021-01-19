const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');
const diff = $('.diff');
const number_input = $('#number');
const export_btn = $('#export-btn');
const first_btn = $('#first-btn');
const previous_btn = $('#previous-btn');
const next_btn = $('#next-btn');
const last_btn = $('#last-btn');

const id = getUrlParam('id');
let note, current_n, max_n;

ajaxGet('/api/note/'+id+'/history', null, that => {
  note = that.response;
  $('#note-id')
    .text('id:'+id)
    .attr('href', '/html/note?id='+id);
  max_n = note.Patches.length;
  number_input.attr('max', max_n);
  gotoHistory(1);
  showHistorySize(note);
}, function() {
  //onloadend
  $('#loading').hide();
});

function showHistorySize(note) {
  const initialValue = 0
  let size = note.Patches.reduce(
    (sum, patch) => { return sum + patch.length }, initialValue);
  size = fileSizeToString(size);
  $('#history-size').text(`共 ${note.Patches.length} 个历史版本，合计 ${size}`);
}

function gotoHistory(n) {
  current_n = n;
  const diffString = note.Patches[n-1];
  const diffJson = Diff2Html.parse(diffString);
  const diffHtml = Diff2Html.html(diffJson, { 
    drawFileList: false,
  });
  diff.html(diffHtml);
}

first_btn.click(() => {
  first_btn.prop('disabled', true);
  previous_btn.prop('disabled', true);
  next_btn.prop('disabled', false);
  last_btn.prop('disabled', false);  
  number_input.val(1);
  gotoHistory(1);
});

last_btn.click(() => {
  next_btn.prop('disabled', true);
  last_btn.prop('disabled', true);
  first_btn.prop('disabled', false);
  previous_btn.prop('disabled', false);  
  number_input.val(max_n);
  gotoHistory(max_n);
});

previous_btn.click(() => {
  if (current_n == max_n) {
    next_btn.prop('disabled', false);
    last_btn.prop('disabled', false);  
  }
  const n = current_n - 1
  if (n == 1) {
    first_btn.prop('disabled', true);
    previous_btn.prop('disabled', true);
  }
  number_input.val(n);
  gotoHistory(n);
});

next_btn.click(() => {
  if (current_n == 1) {
    first_btn.prop('disabled', false);
    previous_btn.prop('disabled', false);  
  }
  const n = current_n + 1
  if (n == max_n) {
    next_btn.prop('disabled', true);
    last_btn.prop('disabled', true);
  }
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
