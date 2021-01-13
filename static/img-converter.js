const sizeLimitElem = $('#size-limit');
const fileInput = $('#file-input');
const convert_btn = $('#convert-btn');
const sizeElem = $('#img-size');
const copy_block = $('#copy-block');
const copy_img = $('#copy-img');
const copy_ref = $('#copy-ref');
const preview = $('#img-preview');

let file, dataURL, img_ref;

fileInput.change(event => {
  if (!event.target.files[0]) return;
  file = event.target.files[0];
});

convert_btn.click(event => {
  event.preventDefault();
  if (!file) {
    insertInfoAlert('请选择文件');
    return;
  }
  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.addEventListener('load', function() {
    drawThumbResize(reader.result).then(img_resized => {
      copy_block.show();
      preview.show().attr('src', img_resized);
      const img_id = 'img' + dayjs().valueOf();
      dataURL = `[${img_id}]:${img_resized}`;
      img_ref = `![${file.name}][${img_id}]`;
      const size = fileSizeToString(img_resized.length);
      $('.alert').remove();
      insertSuccessAlert(`转码成功, size: ${size}`);
    })
    .catch(() => {
      preview.hide();
      copy_block.hide();
      insertErrorAlert('image error');
    });
  }); 
});

const clipboardImg = new ClipboardJS('#copy-img', {
  text: () => { return dataURL; }
});
const clipboardRef = new ClipboardJS('#copy-ref', {
  text: () => { return img_ref; }
});
clipboardImg.on('success', clipOnSuccess);
clipboardImg.on('error', clipOnError);
clipboardRef.on('success', clipOnSuccess);
clipboardRef.on('error', clipOnError);
function clipOnSuccess() {
  insertSuccessAlert('复制成功', copy_block);
}
function clipOnError(e) {
  console.error('Action:', e.action);
  console.error('Trigger:', e.trigger);
  insertErrorAlert('复制失败，详细信息见控制台');
}
async function drawThumbResize(src) {
  const [canvas, changed] = await resizeLimit(src, parseInt(sizeLimitElem.val()));

  // 如果图片不需要缩小，就直接返回src.
  if (!changed) {
    return src;
  }
  const img_resized = canvas.toDataURL('image/jpeg', 0.85);
  return img_resized;
}
  
// ResizeLimit resizes the src if it's long side bigger than limit.
// Use default limit if limit is set to zero or null.
function resizeLimit(src, limit) {
  return new Promise((resolve, reject) => {
    let img = document.createElement('img');
    img.src = src;
    img.onload = function() {
      let [dw, dh] = limitWidthHeight(img.width, img.height, limit);

      // 如果图片小于限制值，其大小就保持不变。
      if (dw == img.naturalWidth && dh == img.naturalHeight) {
        resolve([null, false]);
      }

      let canvas = document.createElement('canvas');
      canvas.width = dw;
      canvas.height = dh;
      let ctx = canvas.getContext('2d');
      ctx.drawImage(img, 0, 0, img.naturalWidth, img.naturalHeight, 0, 0, dw, dh);
      resolve([canvas, true]);
    };
    img.onerror = reject;
  });
}
  
function limitWidthHeight(w, h, limit) {
  if (!limit) {
    limit = 480 // 默认边长上限 480px
  }
  // 先限制宽度
  if (w > limit) {
    h *= limit / w
    w = limit
  }
  // 缩小后的高度仍有可能超过限制，因此要再判断一次
  if (h > limit) {
    w *= limit / h
    h = limit
  }
  return [w, h];
}
  