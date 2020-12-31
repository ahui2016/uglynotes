let isProtected = false;

const id = getUrlParam('id');
const form = new FormData();
form.append('id', id);

ajaxPost(form, '/api/history', null, that => {
  const history = that.response;
  const createdAt = dayjs(history.CreatedAt);
  $('#datetime').text(createdAt.format('YYYY-MM-DD HH:mm:ss'));
  $('#note-id')
    .text('id:'+history.NoteID)
    .attr('href', '/html/note?id='+history.NoteID);
  $('#history-id').text(history.ID);
  $('#size').text(fileSizeToString(history.Size));

  const protected = $('#protected');
  const protectBtn = $('#protect');
  const unprotectBtn = $('#unprotect');
  protectBtn.click(setProtected);
  unprotectBtn.click(setProtected);
  if (history.Protected) protectToggle();

  function protectToggle() {
    isProtected = !isProtected;
    protected.toggle();
    protectBtn.toggle();
    unprotectBtn.toggle();
  }

  function setProtected(event) {
    const form = new FormData();
    form.append("id", history.ID);
    form.append("protected", !isProtected);
    ajaxPut(
        form, '/api/history/protected', $(event.currentTarget), () => {
      protectToggle();
    });
  }

  const plaintext = $('.plaintext.contents');
  plaintext.text(history.Contents);

  const markdown = $('.markdown.contents');
  const dirty = marked(history.Contents);
  const clean = DOMPurify.sanitize(dirty);
  markdown.html(clean);

  $('input[name="note-type"]').change(() => {
    plaintext.toggle();
    markdown.toggle();
  });
  
}, function() {
  //onloadend
  $('#loading').hide();
});
  