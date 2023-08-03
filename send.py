import httpx


headers = {"Authorization": "e14237fe-53e8-4d8f-8a4f-18867c5862f5"}


def submit():
    res = httpx.post("http://localhost:9000/submit", headers=headers, data={
		"mid": "2",
		"time":   "1690442897",
		"text":   "你好李鑫",
		"source": "来自牛魔",
		"platform":    "bilibili",
		"uid":         "434334701",
		"create":   "1690442797",
		"name":        "七海Nana7mi",
		"face":        "https://i2.hdslb.com/bfs/face/f261f5395f1f0082b106f7a23b9424a922b046bb.jpg",
		"description": "大家好，测试这里",
		"follower":    "989643",
		"following":   "551",
		"attachments": ["https://i1.hdslb.com/bfs/face/86faab4844dd2c45870fdafa8f2c9ce7be3e999f.jpg", "https://i2.hdslb.com/bfs/face/f261f5395f1f0082b106f7a23b9424a922b046bb.jpg"]
	})
    print(res.json())


def ping():
    res = httpx.get("http://localhost:9000/user/me", headers=headers)
    print(res.text)


submit()