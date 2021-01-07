ajaxGet('/api/tag/group/all', null, that => {
  if (!that.response) {
    $('#count').text(0);
    return;
  }
  $('#count').text(that.response.length);

  that.response.forEach(group => {
    const updatedAt = dayjs(group.UpdatedAt);
    const item = $('#li-tmpl').contents().clone();
    item.insertAfter('#li-tmpl');

    item.find('.datetime').text(updatedAt.format('YYYY-MM-DD HH:mm:ss'));
    const groupElem = item.find('.group');
    groupElem.attr('href','/html/search?tags=' + encodeURIComponent(addPrefix(group.Tags)));
    group.Tags.forEach(tag => {
      const tagElem = $('#tag-tmpl').contents().clone();
      tagElem
        .text(tag)
        .attr('href', '/html/tag/?name=' + encodeURIComponent(tag));
      tagElem.insertBefore(groupElem);
    });
  });
}, () => {
  // onloadend
  $('#loading').hide();
});
