<p align="center">
  <a href="https://github.com/Drelf2018/webhook/">
    <img src="https://user-images.githubusercontent.com/41439182/220989932-10aeb2f4-9526-4ec5-9991-b5960041be1f.png" height="200" alt="webhook">
  </a>
</p>

<div align="center">

# webhook

_✨ 你说得对，但是 `webhook` 是基于 [weibo-webhook](https://github.com/Drelf2018/weibo-webhook) 改良的分布式博文收集终端 ✨_  


</div>

<p align="center">
  <a href="/">文档</a>
  ·
  <a href="https://github.com/Drelf2018/webhook/releases/">下载</a>
</p>

## 使用

下面是一个简易的提交脚本 ～(∠・ω< )

需要安装前置库 [post-submitter](pypi.org/project/post-submitter) ([Github](https://github.com/Drelf2018/submitter))

```shell
pip install post-submitter
```

```python
from loguru import logger

from submitter import Submitter, Weibo

URL = "http://localhost:9000"
TOKEN = "********"
UID = "188888131"
PostList = []


@Submitter(url=URL, token=TOKEN, dedeuserid=UID)
async def _(sub: Submitter):
    wb = Weibo()
    @sub.job(interval=5, uid=7198559139)
    async def _(uid: int):
        async for post in wb.posts(uid):
            if post.mid not in PostList:
                PostList.append(post.mid)
                logger.info(post)
                err = await sub.submit(post)
                if err is not None:
                    logger.error(err)
                else:
                    logger.info("提交成功")
```
