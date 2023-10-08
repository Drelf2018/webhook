from submitter import *

URL = "http://localhost:9000"
TOKEN = "********"
UID = ""

post = Post(
    mid         = "10086",
    time        = "1690442897",
    text        = "你好李鑫",
    source      = "来自牛魔",
    blogger     = Blogger(
        platform    = "bilibili",
        uid         = "434334701",
        name        = "七海Nana7mi",
        create      = "1690442797",
        follower    = "989643",
        following   = "551",
        description = "虚拟艺人",
        face        = Attachment("https://i2.hdslb.com/bfs/face/f261f5395f1f0082b106f7a23b9424a922b046bb.jpg"),
        pendant     = None,
    ),
    attachments = [],
    repost      = Post(
        mid         = "1000",
        time        = "1690442097",
        text        = "被转发",
        source      = "来自华为Mate60pro",
        blogger     = Blogger(
            platform    = "bilibili",
            uid         = "188888131",
            name        = "脆鲨12138",
            create      = "1690442797",
            follower    = "0",
            following   = "1",
            description = "大家好，测试这里",
            face        = Attachment("https://i1.hdslb.com/bfs/face/86faab4844dd2c45870fdafa8f2c9ce7be3e999f.jpg"),    
            pendant     = None,
        ),
        attachments = [
            Attachment("https://i1.hdslb.com/bfs/face/86faab4844dd2c45870fdafa8f2c9ce7be3e999f.jpg"),
            Attachment("https://i2.hdslb.com/bfs/face/f261f5395f1f0082b106f7a23b9424a922b046bb.jpg"),
        ],
        repost      = None,
        comments    = [],
    ),
    comments = [],
)

@Submitter(url=URL, token=TOKEN, dedeuserid=UID)
async def _(sub: Submitter):
    @sub.job(interval=10)
    async def submit():
        err = await sub.submit(post)
        if isinstance(err, ApiException):
            print("ApiException:", err)
        elif isinstance(err, Exception):
            print("Exception:", err)

    # @sub.job(3, 3)
    # async def query():
    #     posts = await sub.posts(begin="1690442894")
    #     prefix = ">>>"
    #     for p in posts:
    #         print(f"{prefix} Post({p.mid}, {p.name}, {p.text}, {p.date})")
    #         prefix = "   "
    #     print()
