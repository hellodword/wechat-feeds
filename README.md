

## Contribution

感谢 [Treblex](https://github.com/Treblex) 的 [Treblex/wechat-feeds-page](https://github.com/Treblex/wechat-feeds-page), 在他的基础上我将 mdui 换成了 elementui(其实相对他的代码来说, 我这样乱改是代码质量的下降, 但我实在不擅长前端的内容...), 拼拼凑凑姑且实现了一个简单的页面, 希望有擅长前端的朋友贡献代码.

目前是部署在 github pages 上面, 解析用的是 cloudflare, 数据存放在 [feeds 分支的 details.json 文件](https://github.com/hellodword/wechat-feeds/blob/feeds/details.json) 中, [VERSION 文件](https://github.com/hellodword/wechat-feeds/blob/pages/VERSION) 是自动更新的, 也就是 feeds 分支的最新 commit 号, 因为前端是通过 jsdelivr 来获取 `details.json` 的, 如果不指定 commit, 缓存问题很严重, [按照网上的说法强制刷新似乎也不好使](https://blog.juanertu.com/archives/cbcd1946.html).

如果有朋友愿意贡献代码, 这方面咱也不懂, 所以没有什么注意事项, 这就一个展示页面, 只要能继续托管在 github pages 上, 其余可以完全按照你的想法来, 怎么好看怎么酷炫怎么来😂

## TODO 

- [x] jsdelivr 缓存问题, 通过 VERSION 文件解决
- [x] 实现搜索/过滤
- [x] 一次性加载的性能问题，实现无限滚动
