import asyncio
import json
from typing import Any, Tuple, Union

import httpx
from bilibili_api import Credential as _Credential
from bilibili_api import comment, live


class Credential(_Credential):
    def __init__(
        self,
        url: str,
        token: str | None = None,
        sessdata: str | None = None,
        bili_jct: str | None = None,
        dedeuserid: Union[str, None] = None,     
    ) -> None:
        super().__init__(sessdata=sessdata, bili_jct=bili_jct, dedeuserid=dedeuserid)
        self.session = httpx.AsyncClient(base_url=url)
        self.token = token if token is not None else ""

    async def __aenter__(self): return await self.check()

    async def __aexit__(self, type, value, trace): ...

    async def request(self, method, url, *args, **kwargs) -> Tuple[Any, Exception | None]:
        try:
            resp = await self.session.request(method, url, headers={"Authorization": self.token}, *args, **kwargs)
            assert resp.status_code == 200, "请求失败"
            result = resp.json()
            if result["code"] == 0:
                return result["data"], None
            return result, Exception(result["message"])
        except Exception as e:
            return None, e

    async def get(self, url, *args, **kwargs):
        return await self.request("GET", url, *args, **kwargs)
    
    async def post(self, url, *args, **kwargs):
        return await self.request("POST", url, *args, **kwargs)

    async def check(self):
        if self.dedeuserid is None:
            self.dedeuserid = await self.get_uid()
        if not await self.check_token():
            await self.get_token()
        return self

    async def check_token(self) -> bool:
        data, err = await self.get("/me")
        if err is None:
            return data["uid"] == str(self.dedeuserid)
        return False

    async def get_uid(self):
        info = await live.get_self_info(self)
        return str(info["uid"])

    async def get_token(self):
        data, err = await self.get("/token", params={"uid": self.dedeuserid})
        if err is not None:
            raise err
        resp = await self.send_comment(data["token"], data["oid"])
        if resp["success_action"] != 0:
            raise Exception("send comment failed")
        await self.register(data["auth"])

    async def send_comment(self, text: str, oid: int):
        return await comment.send_comment(text, oid, comment.CommentResourceType.DYNAMIC, credential=self)

    async def register(self, auth: str):
        self.token = ""
        for i in range(5):
            await asyncio.sleep(3)
            print(f"register {i}.")
            data, err = await self.get("/register", params={"Authorization": auth})
            if err is None:
                self.token = data
                break

    async def submit(self):
        data = {
            "mid":    "2",
            "time":   "1690442897",
            "text":   "你好李鑫",
            "source": "来自牛魔",
            "platform":    "bilibili",
            "uid":         "434334701",
            "create":      "1690442797",
            "name":        "七海Nana7mi",
            "face":        attachment("https://i2.hdslb.com/bfs/face/f261f5395f1f0082b106f7a23b9424a922b046bb.jpg"),
            "description": "大家好，测试这里",
            "follower":    "989643",
            "following":   "551",
            "attachments": [
                attachment("https://i1.hdslb.com/bfs/face/86faab4844dd2c45870fdafa8f2c9ce7be3e999f.jpg"),
                attachment("https://i2.hdslb.com/bfs/face/f261f5395f1f0082b106f7a23b9424a922b046bb.jpg"),
            ],
            "repost": {},
            "comments": [],
        }
        data, err = await self.post("/submit", data=data)
        print(data, err)


def dict2json(dic: dict = None, **kwargs) -> str:
    if dic is None:
        dic = {}
    dic.update(kwargs)
    return json.dumps(dic)


def attachment(url: str) -> str:
    return dict2json(url=url)


URL = "http://localhost:9000"
TOKEN = "********"
SESSDATA = ""
BILI_JET = ""


async def main():
    async with Credential(url=URL, token=TOKEN, sessdata=SESSDATA, bili_jct=BILI_JET) as cred:
        await cred.submit()


asyncio.run(main())
