# uglynotes 丑丑笔记

一个虽然长得丑，但在功能上有一些特色的笔记软件。


## 重要更新

- 2021-03-01 (change: 前端采用mj.js 和 water.css) 详见 Releases.md
- 2021-02-03 (change: 数据库改用 sqlite)
- 2021-01-25 (add: 快捷键)
- 2021-01-23 (add: 备份/导出)
- 2021-01-20 (upgrade: 历史版本升级)


## 特色一：历史版本完全保留

- 用户只管写，不用点击保存；放心大胆修改文章，修改过程会以 “历史版本” 的形式被记录，随时可以找回。
- 默认每隔 5 分钟产生一个历史版本，该间隔时间可以自由设置。
- 每个历史版本**只保存**与上一个版本之间的差异部分，因此不会占用太多储存空间。
- 在历史版本页面提供 “上一个”和“下一个” 按钮，可非常方便、直观地查看每个版本的变化（高亮显示变化位置）。


## 特色二：Markdown 内嵌图片

- 特色一历史版本功能是我在逛 V2EX 的时候看大家讨论笔记软件时获得的灵感，这个特色二也同样是看到有人说起，我才想到做这个功能。
- 目前我做了一个独立的页面 http://note.ai42.xyz/converter 用来将图片转码，转码后可粘贴到任何 markdown 文件中（不局限于本站，任何支持 markdown 的地方都可以使用），即可让图片直接内嵌在 markdown 文件里。
- 不需要图床，因此也不用怕图床失效，图片数据就在 markdown 文件里，因此图片永远有效。


## 特色三：对标签管理的重新思考

- 用标签来管理文件，比使用文件夹更科学、更方便好用。但为了照顾用户习惯，很多支持标签的软件都会同时支持文件夹。
- 但习惯的力量是可怕的，一旦支持文件夹，用户就不会认真对待标签，结果有的文件设了标签，同时有很多文件只是归类到文件夹里，完全没有标签。
- 对于标签管理系统来说，“有的文件有标签，有的文件没有标签” 是灾难性的，这导致整个标签系统名存实亡，彻底沦有一个可有可无的辅助角色，发挥不出应有的效果。
- 因此，在本软件中尝试不使用文件夹，只使用标签，并且要求每篇笔记至少要有两个标签。同时提出了一个 “标签组” 的概念，最大限度发挥标签的效果。


## Tag Group (标签组)

- 每个标签组要求至少包含 2 个标签。
- 在标签组列表页面，可通过标签组来搜索笔记，也可通过标签组来创建新笔记。
- 原则上只能通过标签组来创建新笔记，这是为了改变用户习惯，确保每篇笔记都有标签。
- 下面通过一些例子来说明这样做的好处。

### 标签组示例

我们写笔记，最大的目的是为了日后能轻松找出笔记。在有了 “标签组” 这个高效率管理工具之后，只需要遵从一个简单的原则，即可轻松创建出非常有利于检索的标签组：

**原则：一两个共性标签 + 一两个唯一性标签**

比如：

- `#editor` `#emacs` `#快捷键`
- `#editor` `#emacs` `#org-mode`
- `#editor` `#vim` `#快捷键`
- `#editor` `#vim` `#vimrc`
- `#操作系统` `#Windows` `#快捷键`

当我们用上述标签组来创建一些笔记后，

- 搜 `vim`(共性标签) 能找出与 vim 有关的快捷键、vimrc等笔记
- 搜 `org-mode` 或 `vimrc`(唯一性标签) 即可直接找出最精确的结果
- 搜 `editor`(更大范围的共性标签) 又能扩大搜索范围
- 还可以搜 `editor` + `快捷键` 来找全部编辑器的快捷键而不被 `操作系统` 的快捷键污染搜索结果

可见，标签管理很科学，也很易用，我们以前不这样用，是因为在有文件夹的系统里有大量文件没有标签，导致我们每当想通过标签来搜索文件时都心里没底，总觉得有漏网之鱼。

在规定必须使用标签的系统里，我们可以体验标签管理的真正实力。


## demo 演示版

- http://note.ai42.xyz (密码 abc)
- demo 的笔记字数限制、数据库总容量、产生历史版本的间隔时间、自动提交次数上限等，都设置了比较低的数值，实际使用正式版时这些数值都可以自由设置。
- demo 服务器在美国，一个非常低配的 VPS, 因此访问速度比较慢，这是网络问题不是程序问题。


## 关于丑

- 我对前端界面美化实在不擅长，折腾起来太花时间，就索性不折腾了。没有用任何前端框架，CSS 也尽可能少用，因此是很原始的风格。
- 主要考虑桌面屏幕，没有做手机屏幕适配，但我试了一下手机使用也……勉强能用。
- 大体上是前后端分离的，后端只向前端返回 json, 从不返回渲染过的网页，因此有前端能力的朋友们可以轻松改写前端页面。


## 安装运行

```
$ cd ~
$ git clone https://github.com/ahui2016/uglynotes.git 
$ cd uglynotes && go build
$ ./uglynotes &
```

### 密码等的设置

绝大部分设置（包括密码）都汇总在 settings.json 文件里，请用文本编辑器打开该文件，修改设置，修改后需要重启程序：

```
$ killall uglynotes
$ ./uglynotes &
```

启动程序时默认在当前目录下找 settings.json, 但也可以用 -config 参数指定配置文件：

```
$ ./uglynotes -config /path/to/another_settings.json &
```

关于 settings.json 里各项目的详细说明请看 settings/settings.go

另外，“创建历史版本的间隔时间” 和 “自动保存（自动更新）次数的上限” 在 public/util.js 中设置，修改后不需要重启程序，而是需要在浏览器用 ctrl+shift+R 强制刷新。

### 数据库文件夹的设置

- 可使用 -dir 参数指定数据库文件夹：

```
$ ./uglynotes -dir /path/to/db-folder
```

- 如果不使用 -dir, 则默认在用户目录 ($HOME) 下创建数据库文件夹，文件夹名称可在 settings.json 中设置。


## 备份/数据导出

- 整个数据库只有一个文件 uglynotes.db (软件启动时会在终端打印该文件的位置)，因此只要备份这个文件就可以了。
- 另外，在 Backup 页面可将数据库中的全部笔记导出为 uglynotes.json。