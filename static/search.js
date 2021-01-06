const search_input = $('#search-input');
const search_btn = $('#search-btn');
const loading = $('#loading');

let searchFor = 'tags';

search_btn.click(event => {
  event.preventDefault();
  const pattern = search_input.val().trim();
  if (pattern == '') {
    insertInfoAlert('请输入搜索内容');
    search_input.focus();
    return;
  }

  if (searchFor == 'tags') {
    searchTags();
  }
});

function searchTags() {
  // getTags 返回标签集合， addPrefix 把集合数组转化为字符串。
  const tagSet = getTags(search_input);
  const tags = addPrefix(tagSet);
  const url = '/api/search/tags/' + encodeURIComponent(tags);
  loading.text('searching: ' + addPrefix(tagSet, '#'));
  ajaxGet(url, search_btn, that => {
    console.log(that.response);
    insertSuccessAlert('ok');
  });

}