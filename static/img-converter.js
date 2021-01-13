const preview = $('#img-preview');
const fileInput = $('#file-input');
const sizeElem = $('#img-size');
const copy_block = $('#copy-block');
const copy_btn = $('#copy');

let dataURL;

fileInput.change(event => {
  const file = event.target.files[0];
  if (!file) return;

  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.addEventListener('load', function() {
    drawThumbResize(reader.result).then(img_resized => {
      copy_block.show();
      preview.show().attr('src', img_resized);
      dataURL = `![${file.name}][${file.name}]\n\n[${file.name}]:${img_resized}`;
      const size = fileSizeToString(img_resized.length);
      insertSuccessAlert(`转码成功, size: ${size}`);
    })
    .catch(() => {
      preview.hide();
      copy_block.hide();
      insertErrorAlert('image error');
    });
  });    
});

const clipboard = new ClipboardJS('#copy', {
  text: () => { return dataURL; }
});
clipboard.on('success', () => {
  insertSuccessAlert('已复制，请粘贴到 markdown 文件中', copy_btn);
});
clipboard.on('error', e => {
  console.error('Action:', e.action);
  console.error('Trigger:', e.trigger);
  insertErrorAlert('复制失败，详细信息见控制台');
});

async function drawThumbResize(src) {
  const [canvas, changed] = await resizeLimit(src, null);

  // 如果图片不需要缩小，就直接返回src.
  if (!changed) {
    return src;
  }
  const img_resized = canvas.toDataURL('image/jpeg');
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
    limit = 600 // 默认边长上限 600px
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
  