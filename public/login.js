const pw_input = $('#password');
const submit_btn = $('#submit');

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
    submit_btn.prop('disabled', true);
    window.location.reload();
  });
});
