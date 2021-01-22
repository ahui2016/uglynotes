const loading = $('#loading');
const pw_input = $('#password');
const submit_btn = $('#submit');
const formElem = $('form');

ajaxGet('/check', null, that => {
  if (that.response.message == "OK") {
    insertSuccessAlert('已登入')
    navi.show();
    return;
  }
  formElem.show();
  pw_input.focus();
}, function() {
  // onloadend
  loading.hide();
});

submit_btn.click(event => {
  event.preventDefault();
  let password = pw_input.val();
  if (password == '') {
    insertInfoAlert('请输入密码');
    pw_input.focus();
    return;
  }
  
  let form = new FormData();
  form.append('password', password);

  ajaxPost(form, '/login', submit_btn, function() {
    $('.alert').remove();
    submit_btn.prop('disabled', true);
    window.location = '/home';
  });
});

function restoretags() {
  ajaxGet('/reset-all-tags', null, ()=>{console.log('OK')});
}