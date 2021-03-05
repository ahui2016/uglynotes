# 数据备份、导出等高级功能

## 数据备份、导出

在 Backup 页面可下载整个数据库文件(sqlite), 也可以将全部数据导出为 JSON 格式。

## 导入 JSON

在 Backup 页面导出的 JSON 文件是原始数据，将其重新导入数据库或转换为一篇篇笔记的功能还没有做，今后会做的。

## 自动提交频率

- 在编辑笔记的页面，默认每隔 5 分钟自动向服务器提交一次，同时产生一个历史版本。
- 但有时可能会希望每隔 30 秒就自动提交一次，有时又可能会想每隔 30 分钟才提交一次，这个频率是可以调整的。
- 按 F12 打开控制台，执行命令 restartAutoSubmit(delay) 即可重新设置间隔时间，单位是秒，比如 `restartAutoSubmit(60 * 3)` 可设置为每隔 3 分钟自动提交一次。
- 而执行 `stopAutoSubmit()` 可完全停止自动提交。
- 该方法只对当前页面有效，刷新页面或编辑另一篇笔记会恢复默认频率。

## 快捷键

在编辑笔记的页面可使用以下快捷键：

- **Alt-Shift-P**: 预览(preview)
- **Alt-Shift-E**: 编辑(edit)
- **Alt-Shift-C**: 笔记编辑框(contents)
- **Alt-Shift-T**: 标签编辑框(tags)

在 Chrome 里可以使用 **Alt-C** 和 **Alt-T** 分别跳到笔记编辑框和标签编辑框。 Mac 用户请用 option 键代替 alt 键。

另外，当光标在笔记编辑框时，按 Tab 键可跳到标签编辑框，再按 Shift + Tab 可跳回到笔记编辑框。
