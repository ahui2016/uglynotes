ajaxGet('/notes/all', null, function(that){
  $('#loading').hide();
  that.response.forEach(note => {
    let updatedAt = dayjs(note.UpdatedAt);
    let item = $('#li-tmpl').contents().clone();
    item.find('.id').text(note.ID);
    item.find('.datetime').text(updatedAt.format('MMM D, HH:mm'));
    item.find('.title')
      .attr('href', '/note?id='+note.ID)
      .text(note.Title);
    item.find('.tags').text(addPrefix(note.Tags, '#'));
    item.insertAfter('#li-tmpl');
  });
});
