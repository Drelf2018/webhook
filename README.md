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

## 简介

本库接口使用 `RESTful` 风格，返回值为 `JSON` 对象，结构如下：

```go
type Response struct {
	Code  int    `json:"code"`
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}
```

其中 `Code` 为业务码，值为 `0` 表示业务正常，其余值均代表不同错误码。

其中 `Error` 为错误提示信息，业务正常时无此字段。

其中 `Data` 为返回数据，可能为空。

## 模型

这里提到的模型可以在 [model](https://github.com/Drelf2018/webhook/tree/main/model) 文件夹找到。

### 博文

```go
type Blog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"` // 数据库内序号
	CreatedAt time.Time `json:"created_at"`                         // 数据库内创建时间

	Submitter string `json:"submitter" gorm:"index:idx_blogs_query,priority:2"`      // 提交者
	Platform  string `json:"platform" gorm:"index:idx_blogs_query,priority:5"`       // 发布平台
	Type      string `json:"type" gorm:"index:idx_blogs_query,priority:4"`           // 博文类型
	UID       string `json:"uid" gorm:"index:idx_blogs_query,priority:3"`            // 账户序号
	MID       string `json:"mid" gorm:"index:idx_blogs_query,priority:1;column:mid"` // 博文序号

	URL    string    `json:"url"`    // 博文网址
	Text   string    `json:"text"`   // 文本内容
	Time   time.Time `json:"time"`   // 发送时间
	Title  string    `json:"title"`  // 文章标题
	Source string    `json:"source"` // 博文来源
	Edited bool      `json:"edited"` // 是否编辑

	Name        string `json:"name"`        // 账户昵称
	Avatar      string `json:"avatar"`      // 头像网址
	Follower    string `json:"follower"`    // 粉丝数量
	Following   string `json:"following"`   // 关注数量
	Description string `json:"description"` // 个人简介

	ReplyID   *uint64 `json:"reply_id"`                             // 被本文回复的博文序号
	Reply     *Blog   `json:"reply"`                                // 被本文回复的博文
	CommentID *uint64 `json:"comment_id"`                           // 被本文评论的博文序号
	Comments  []Blog  `json:"comments" gorm:"foreignKey:CommentID"` // 本文的评论

	Assets pq.StringArray `json:"assets" gorm:"type:text[]"`    // 资源网址
	Banner pq.StringArray `json:"banner" gorm:"type:text[]"`    // 头图网址
	Extra  map[string]any `json:"extra" gorm:"serializer:json"` // 预留项
}
```

### 任务

任务公开后，就可以被搜索到，除非你希望教他人如何创建任务，且你的请求参数中不包含敏感信息，否则不要公开。

任务启用后，每当一个新博文被提交后，就会通过筛选条件判断是否要运行任务。

```go
type Task struct {
	ID        uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Public bool   `json:"public"`  // 是否公开
	Enable bool   `json:"enable"`  // 是否启用
	Name   string `json:"name"`    // 任务名称
	Icon   string `json:"icon"`    // 任务图标
	Method string `json:"method"`  // 请求方法
	URL    string `json:"url"`     // 请求地址
	Body   string `json:"body"`    // 请求内容
	Header Header `json:"header"`  // 请求头部
	README string `json:"README"`  // 任务描述
	ForkID uint64 `json:"fork_id"` // 复刻来源

	ForkCount int `json:"fork_count" gorm:"-"` // 被复刻次数

	Filters []Filter     `json:"filters"` // 筛选条件
	Logs    []RequestLog `json:"logs"`    // 请求记录
	UserID  string       `json:"user_id"` // 外键
}
```

### 筛选条件

博文筛选条件，用来描述一类博文，例如：`filter1` 表示所有平台为 `"weibo"`、类型为 `"comment"` 的博文，`filter2` 表示所有由 `"114"` 提交的用户 `"514"` 的博文。

```go
type Filter struct {
	Submitter string `json:"submitter" form:"submitter"` // 提交者
	Platform  string `json:"platform" form:"platform"`   // 发布平台
	Type      string `json:"type" form:"type"`           // 博文类型
	UID       string `json:"uid" form:"uid"`             // 账户序号
	TaskID    uint64 `json:"-" form:"-"`                 // 外键
}

var filter1 = Filter{
	Platform: "weibo",
	Type: "comment",
}

var filter2 = Filter{
	Submitter: "114",
	UID: "514",
}
```

### 请求记录

每个任务运行后，都会留下一条记录，保存运行任务的相应内容或者发生的错误。

```go
type RequestLog struct {
	BlogID     uint64    `json:"blog_id"`
	CreatedAt  time.Time `json:"created_at"`
	FinishedAt time.Time `json:"finished_at"`
	Result     any       `json:"result" gorm:"serializer:json"` // 响应为 JSON 会自动解析
	Error      string    `json:"error"`                         // 请求过程中发生的错误
	TaskID     uint64    `json:"-" gorm:"index:idx_logs_query"` // 外键
}
```

### 查询条件

特殊的，在文件 [visitor.go](https://github.com/Drelf2018/webhook/tree/main/api/visitor.go) 中有这样一个结构体，允许从后端获取博文时使用部分 `SQL` 参数，参数 `Reply` 为真时会将博文转发的那条博文一并返回。参数 `Comments` 为真时会将博文的评论一并返回。

```go
type Condition struct {
	// 是否包含转发 默认 true
	Reply bool `json:"reply" form:"reply"`

	// 是否包含评论 默认 true
	Comments bool `json:"comments" form:"comments"`

	// 查询排列顺序 默认 time desc
	Order string `json:"order" form:"order"`

	// 查询行数 默认 30
	Limit int `json:"limit" form:"limit"`

	// 查询偏移 默认 0
	Offset int `json:"offset" form:"offset"`

	// 其他条件
	Conds []string `json:"conds" form:"conds"`
}
```

## 接口权限

接口权限分为访客、用户、管理员和所有者。

- 访客：所有人都可以使用的接口

- 用户：通过了 `JWT` 校验的用户可以使用的接口

- 管理员：在配置文件中记录或所有者后期指定的用户可以使用的接口

- 所有者：仅配置文件中记录的所有者可以使用的接口

### 校验

本库采用 `JWT` 校验。在完成注册后，可以通过接口获取或刷新用户 `Token`。

在请求头中使用：

```
Authorization: $Token
```

或在请求参数中使用：

```
auth=$Token
```

即可进行校验。

### PATCH 请求

对于一个 PATCH 接口，需要以 `JSON` 格式发送包含操作对象的列表的请求体，单条操作对象包含以下字段：

| 字段名 | 字段类型 | 字段含义                                                     |
| ------ | -------- | ------------------------------------------------------------ |
| op     | string   | 操作类型，一般使用 `replace` 表示替换、`add` 表示添加、`remove` 表示移除 |
| path   | string   | 操作路径，一般表示要操作的资源                               |
| value  | string   | 操作数据，一般表示替换、添加操作时的新值                     |

```json
[
  {
    "op": "replace",
    "path": "/nickname",
    "value": "用户21452505"
  },
  {
    "op": "add",
    "path": "/ban",
    "value": "1000"
  }
]
```

在这个例子中，第一条操作含义：

- `op` 是 `replace`，表示我们要替换数据。
- `path` 是 `/nickname`，表示我们要替换的是昵称。
- `value` 是 `用户21452505`，是我们想要设置的新昵称。

第二条操作含义：

- `op` 是 `add`，表示我们要添加数据。
- `path` 是 `/ban`，表示我们要添加的是封禁时间。
- `value` 是 `1000`，是我们想要添加 `1000` 毫秒的封禁时间。

## 接口列表

| 请求方法 | 请求路径            | 接口介绍       | 接口权限 |
| -------- | ------------------- | -------------- | -------- |
| GET      | /public/*filepath   | 获取资源       | 访客     |
| ANY      | /forward/*url       | 请求转发       | 访客     |
| GET      | /api/version        | 获取版本信息   | 访客     |
| GET      | /api/valid          | 校验鉴权码     | 访客     |
| GET      | /api/ping           | 更新在线时间   | 访客     |
| GET      | /api/online         | 获取在线状态   | 访客     |
| GET      | /api/token          | 获取 Token     | 访客     |
| GET      | /api/blogs          | 获取博文       | 访客     |
| POST     | /api/blogs          | 获取筛选后博文 | 访客     |
| GET      | /api/blog/:id       | 获取单条博文   | 访客     |
| GET      | /api/tasks          | 获取任务集     | 访客     |
| POST     | /api/user           | 注册账户       | 访客     |
| GET      | /api/user/:uid      | 获取用户信息   | 访客     |
| GET      | /api/user           | 获取自身信息   | 用户     |
| PATCH    | /api/user/:uid      | 修改用户信息   | 用户     |
| GET      | /api/following      | 获取关注的博文 | 用户     |
| POST     | /api/blog           | 提交博文       | 用户     |
| POST     | /api/task           | 提交任务       | 用户     |
| GET      | /api/task/:id       | 获取任务       | 用户     |
| PATCH    | /api/task/:id       | 修改任务       | 用户     |
| DELETE   | /api/task/:id       | 移除任务       | 用户     |
| POST     | /api/test           | 测试任务       | 用户     |
| GET      | /api/logs/*filepath | 获取日志       | 管理员   |
| GET      | /api/root/*filepath | 获取文件       | 所有者   |
| POST     | /api/upload         | 上传文件       | 所有者   |
| GET      | /api/execute        | 执行命令       | 所有者   |
| GET      | /api/shutdown       | 优雅关机       | 所有者   |

### GET /public/*filepath

可以获取服务器已保存的资源，也可以获取一个链接对应的资源。对于链接：

```
https://example.com/resource/image.png?width=200
```

需要去除 `:/` 并将 `?` 转义成 `%3F` 后发起请求：

```
/public/https/example.com/resource/image.png%3Fwidth=200
```

当此请求正常返回后，可以使用以下接口获取已保存的资源：

```
/public/example.com/resource/image.png
```

### ANY /forward/*url

参考 [gin 请求转发](https://blog.csdn.net/qq_29799655/article/details/113841064) 实现的请求转发功能，用来绕过浏览器 `CORS` 策略限制。对于链接：

```
https://api.bilibili.com/x/web-interface/card?mid=2
```

需要去除 `:/` 后发起请求：

```
/forward/https/api.bilibili.com/x/web-interface/card?mid=2
```

### GET /api/version

获取后端的接口版本、环境版本、运行时间、主页文件名信息。

```json
{
    "code": 0,
    "data": {
        "api": "v0.19.0",
        "env": "go1.23.4 windows/amd64",
        "start": "2025-04-18T11:39:39.7941343+08:00",
        "index": null
    }
}
```

### GET /api/valid

按照上文 `JWT` 校验中两种方法提供 `Token` ，后端进行校验并返回真假值。

```json
{
    "code": 0,
    "data": false
}
```

### GET /api/ping

校验通过后，后端记录当前时间作为用户最后一次在线时间。

```json
{
    "code": 0,
    "data": "pong"
}
```

### GET /api/online

获取所有用户最后一次在线距离当前时间的差值，单位毫秒。

```json
{
    "code": 0,
    "data": {
        "188888131": 49826
    }
}
```

### GET /api/token

获取或刷新 `Token` ，请求参数包括：

| 参数名  | 参数类型 | 参数含义         |
| ------- | -------- | ---------------- |
| uid     | string   | 账号             |
| pwd     | string   | 密码             |
| refresh | bool     | 是否刷新，默认否 |

当请求参数 `uid` 为空时，会从请求头中使用 `BasicAuth` 获取账号、密码。这样相比在请求参数中写入账号、密码更加安全，但也更不方便。

```
/api/token?uid=188888131&pwd=123456
```

```json
{
    "code": 0,
    "data": "xxx"
}
```

### GET /api/blogs

获取博文，请求参数包括：

| 参数名  | 参数类型               | 参数含义                                                     |
| ------- | ---------------------- | ------------------------------------------------------------ |
|         | [Condition](#查询条件) | 用 Condition 结构体的 6 个字段作为参数                       |
|         | [Filter](#筛选条件)    | 用 Filter 结构体的 4 个字段作为参数                          |
| mid     | string                 | 要获取博文的序号                                             |
| task_id | []uint64               | 此参数非空时，会忽略 `Filter` 和 `mid` 参数，并从所有公开或本人的任务中匹配任务序号，再合并所有匹配成功的任务的筛选条件，最后用这些条件筛选出博文 |

假设仅存在序号 `1` 的任务，对于链接：

```
/api/blogs?reply=false&limit=3&offset=17&task_id=1&task_id=2
```

会用任务 `1` 筛选出按时间排序下第 `17` 条博文之后的 `3` 条博文，并且不会附带这些博文转发的博文（如果存在）

```json
{
    "code": 0,
    "data": [
        {
            "id": 1576,
            "created_at": "2025-04-13T20:53:16.531681746+08:00",
            "submitter": "188888131",
            "platform": "weibo",
            "type": "blog",
            "uid": "7198559139",
            "mid": "5155073610487050",
            "url": "https://m.weibo.cn/status/Pn6Ip22HE",
            "text": "好大好亮今天！ ",
            "time": "2025-04-13T20:53:08+08:00",
            "title": "",
            "source": "发布于 上海",
            "edited": false,
            "name": "七海Nana7mi",
            "avatar": "https://wx4.sinaimg.cn/orj480/007Raq4zly8i00p49h5kkj30p70p7mzr.jpg",
            "follower": "103.6万",
            "following": "212",
            "description": "蓝色饭团",
            "reply_id": null,
            "reply": null,
            "comment_id": null,
            "comments": [],
            "assets": [
                "https://wx1.sinaimg.cn/large/007Raq4zgy1i0fh4oemz9j32782xn4qq.jpg"
            ],
            "banner": [
                "https://wx4.sinaimg.cn/crop.0.0.640.640.640/007Raq4zgy1hzo2u3qofwj30rs0rsdja.jpg"
            ],
            "extra": {
                "is_top": false,
                "source": "🦈iPhone 14 Pro Max"
            }
        },
        ...
    ]
}
```

### POST /api/blogs

获取筛选后博文，使用 `JSON` 格式发送请求体：

| 参数名  | 参数类型               | 参数含义                               |
| ------- | ---------------------- | -------------------------------------- |
|         | [Condition](#查询条件) | 用 Condition 结构体的 6 个字段作为参数 |
| filters | [[]Filter](#筛选条件)  | 筛选条件                               |

返回格式同基础查询。

### GET /api/blog/:id

获取单条博文，路径参数为博文序号。

```
/api/blog/1576
```

```json
{
    "code": 0,
    "data": {
        "id": 1576,
        "created_at": "2025-04-13T20:53:16.531681746+08:00",
        "submitter": "188888131",
        "platform": "weibo",
        "type": "blog",
        "uid": "7198559139",
        "mid": "5155073610487050",
        "url": "https://m.weibo.cn/status/Pn6Ip22HE",
        "text": "好大好亮今天！ ",
        "time": "2025-04-13T20:53:08+08:00",
        "title": "",
        "source": "发布于 上海",
        "edited": false,
        "name": "七海Nana7mi",
        "avatar": "https://wx4.sinaimg.cn/orj480/007Raq4zly8i00p49h5kkj30p70p7mzr.jpg",
        "follower": "103.6万",
        "following": "212",
        "description": "蓝色饭团",
        "reply_id": null,
        "reply": null,
        "comment_id": null,
        "comments": [],
        "assets": [
            "https://wx1.sinaimg.cn/large/007Raq4zgy1i0fh4oemz9j32782xn4qq.jpg"
        ],
        "banner": [
            "https://wx4.sinaimg.cn/crop.0.0.640.640.640/007Raq4zgy1hzo2u3qofwj30rs0rsdja.jpg"
        ],
        "extra": {
            "is_top": false,
            "source": "🦈iPhone 14 Pro Max"
        }
    }
}
```

### GET /api/tasks

获取任务集，请求参数包括：

| 参数名 | 参数类型 | 参数含义                         |
| ------ | -------- | -------------------------------- |
| key    | string   | 关键词，用来在任务名和描述中搜索 |
| limit  | int      | 查询行数                         |
| offset | int      | 查询偏移                         |

```
/api/tasks?key=图
```

```json
{
    "code": 0,
    "data": [
        {
            "id": 1,
            "created_at": "2025-01-01T10:36:08.250177945Z",
            "public": true,
            "enable": true,
            "name": "微博转图",
            "icon": "",
            "method": "POST",
            "url": "xxx",
            "body": "{{ json . }}",
            "header": {},
            "README": "接收所有微博",
            "fork_id": 0,
            "fork_count": 0,
            "filters": [
                {
                    "submitter": "188888131",
                    "platform": "weibo",
                    "type": "",
                    "uid": ""
                }
            ],
            "logs": null,
            "user_id": "188888131"
        },
        ...
    ]
}
```

### POST /api/user

注册账户，具体请求方式由后端使用的注册函数决定。

### GET /api/user/:uid

获取指定用户信息，路径参数为用户序号，结果不包含任务 `Tasks` 信息。

```
/api/user/188888131
```

```json
{
    "code": 0,
    "data": {
        "uid": "188888131",
        "created_at": "2025-01-01T10:34:44.483175376Z",
        "ban": "0001-01-01T00:00:00Z",
        "role": 4,
        "name": "脆鲨12138",
        "nickname": "",
        "tasks": null
    }
}
```

### GET /api/user

鉴权通过后，获取自身用户信息，结果包含任务 `Tasks` 信息。

```json
{
    "code": 0,
    "data": {
        "uid": "188888131",
        "created_at": "2025-01-01T10:34:44.483175376Z",
        "ban": "0001-01-01T00:00:00Z",
        "role": 4,
        "name": "脆鲨12138",
        "nickname": "",
        "tasks": [
            {
                "id": 1,
                "created_at": "2025-01-01T10:36:08.250177945Z",
                "public": false,
                "enable": true,
                "name": "微博转图",
                "icon": "",
                "method": "POST",
                "url": "xxx",
                "body": "{{ json . }}",
                "header": {},
                "README": "接收所有微博",
                "fork_id": 0,
                "fork_count": 0,
                "filters": [
                    {
                        "submitter": "188888131",
                        "platform": "weibo",
                        "type": "",
                        "uid": ""
                    }
                ],
                "logs": null,
                "user_id": "188888131"
            },
            ...
        ]
    }
}
```

### PATCH /api/user/:uid

鉴权成功后，修改指定用户信息。每个用户均可修改自身部分信息，管理者可以修改权限低于自身的用户的部分信息，例如：昵称、封禁时间等。

用户可选操作包括：

| 操作路径  | 操作类型 | 操作含义       |
| --------- | -------- | -------------- |
| /nickname | replace  | 修改自己的昵称 |

管理者可选操作包括：

| 操作路径  | 操作类型 | 操作含义                                                     |
| --------- | -------- | ------------------------------------------------------------ |
| /nickname | replace  | 修改自己或权限低于自己的用户的昵称                           |
| /name     | replace  | 修改自己或权限低于自己的用户的用户名                         |
| /role     | replace  | 修改权限低于自己的用户的权限，数值大于 `0` 且小于自身权限等级 |
| /ban      | replace  | 修改权限低于自己的用户的封禁时间，参数采用 `RFC3339` 时间格式 |
|           | add      | 添加权限低于自己的用户的封禁时间，参数以毫秒为单位的数字字符串 |
|           | remove   | 移除权限低于自己的用户的封禁时间                             |

```
/api/user/188888131
```

```json
{
    "code": 0,
    "data": "success"
}
```

### GET /api/following

鉴权成功后，获取自己的任务能匹配的博文，请求参数包括：

| 参数名 | 参数类型               | 参数含义                               |
| ------ | -------- | -------------------------------- |
|        | [Condition](#查询条件) | 用 Condition 结构体的 6 个字段作为参数 |

```
/api/following?limit=100&offset=200
```

```json
{
    "code": 0,
    "data": [
        {
            "id": 1276,
            "created_at": "2025-03-19T03:08:52.835047151Z",
            "submitter": "188888131",
            "platform": "weibo",
            "type": "blog",
            "uid": "7198559139",
            "mid": "5145866868364764",
            "url": "https://m.weibo.cn/status/PjfcOz63y",
            "text": "太好听噜 大早上给我听醉了[种树] ",
            "time": "2025-03-19T11:08:50+08:00",
            "title": "",
            "source": "发布于 上海",
            "edited": false,
            "name": "七海Nana7mi",
            "avatar": "https://wx1.sinaimg.cn/orj480/007Raq4zly8hzk8uj9qc8j30e80e8mya.jpg",
            "follower": "103.6万",
            "following": "212",
            "description": "蓝色饭团",
            "reply_id": null,
            "reply": null,
            "comment_id": null,
            "comments": [],
            "assets": [
                "https://wx2.sinaimg.cn/large/007Raq4zgy1hzm3r1qr85j30zu25o7wh.jpg"
            ],
            "banner": [
                "https://wx4.sinaimg.cn/crop.0.0.640.640.640/007Raq4zgy1htmdatc6nyj30u00u0771.jpg"
            ],
            "extra": {
                "is_top": false,
                "source": "🦈iPhone 14 Pro Max"
            }
        },
        ...
    ]
}
```

### POST /api/blog

鉴权成功后，提交博文，使用 `JSON` 格式发送 `Blog` 的请求体，返回值为提交的博文的序号：

| 参数名  | 参数类型 | 参数含义                             |
| ------- | -------- | ------------------------------------ |
|     | [Blog](#博文) | 待提交博文                      |

```json
{
    "code": 0,
    "data": 1276
}
```

### POST /api/task

鉴权成功后，提交任务，使用 `JSON` 格式发送 `Task` 的请求体，返回值为提交的任务的序号：

| 参数名  | 参数类型 | 参数含义                             |
| ------- | -------- | ------------------------------------ |
|     | [Task](#任务) | 待提交任务              |

```json
{
    "code": 0,
    "data": 1
}
```

### GET /api/task/:id

鉴权成功后，获取公开或自己的任务信息，路径参数为任务序号，请求参数包括：

| 参数名 | 参数类型 | 参数含义                 |
| ------ | -------- | ------------------------ |
| limit  | int      | 该任务请求记录的查询行数 |
| offset | int      | 该任务请求记录的查询偏移 |

```
/api/task/1?offset=1
```

```json
{
    "code": 0,
    "data": {
        "id": 1,
        "created_at": "2025-01-01T10:36:08.250177945Z",
        "public": true,
        "enable": false,
        "name": "微博转图",
        "icon": "",
        "method": "POST",
        "url": "xxx",
        "body": "{{ json . }}",
        "header": {},
        "README": "接收所有微博",
        "fork_id": 0,
        "fork_count": 0,
        "filters": [
            {
                "submitter": "188888131",
                "platform": "weibo",
                "type": "",
                "uid": ""
            }
        ],
        "logs": [
            {
                "blog_id": 1599,
                "created_at": "2025-04-16T18:27:26.033047428+08:00",
                "finished_at": "2025-04-16T18:28:12.298975451+08:00",
                "result": "success",
                "error": ""
            },
            ...
        ],
        "user_id": "188888131"
    }
}
```

### PATCH /api/task/:id

鉴权成功后，修改自己的任务，路径参数为任务序号。可选操作包括：

| 操作路径 | 操作类型 | 操作含义                                                     |
| -------- | -------- | ------------------------------------------------------------ |
| /public  | add      | 将任务公开，公开后不可取消公开                               |
| /enable  | replace  | 修改任务是否启用，参数为真假值                               |
| /name    | replace  | 修改任务名称                                                 |
| /icon    | replace  | 修改任务图标链接                                             |
| /method  | replace  | 修改任务请求方法                                             |
| /url     | replace  | 修改任务请求地址                                             |
| /body    | replace  | 修改任务请求体                                               |
| /header  | replace  | 修改任务请求头，参数为 `JSON` 格式的 `map[string]string`     |
| /readme  | replace  | 修改任务描述                                                 |
| /filters | replace  | 修改任务筛选条件，参数为 `JSON` 格式的 [Filter](#筛选条件) 列表 |

```
/api/task/1
```

```json
{
    "code": 0,
    "data": "success"
}
```

### DELETE /api/task/:id

鉴权成功后，删除指定任务，路径参数为任务序号。

```
/api/task/1
```

```json
{
    "code": 0,
    "data": "success"
}
```

### POST /api/test

鉴权成功后，测试已保存或请求体内上传的任务，使用 `JSON` 格式发送请求体：

| 参数名  | 参数类型 | 参数含义                             |
| ------- | -------- | ------------------------------------ |
| blog    | [Blog](#博文) | 测试用博文                           |
| task    | [Task](#任务) | 待测试任务                           |
| blog_id | uint64            | 测试用博文序号，优先于上传的博文使用 |
| task_id | []uint64          | 待测试博文序号，优先于上传的任务使用 |

```python
body = {
    "blog_id": 1276,
    "task": {
        "method": "POST",
        "url": "https://httpbin.org/post",
        "body": "{{ json . }}",
        "header": {"Authorization": "abc123"},
    },
}
```

```json
{
    "code": 0,
    "data": [
        {
            "blog_id": 1276,
            "created_at": "2025-04-28T12:14:01.8963383+08:00",
            "finished_at": "2025-04-28T12:14:02.3443411+08:00",
            "result": {
                "args": {},
                "data": "",
                "files": {},
                "form": {},
                "headers": {
                    "Accept-Encoding": "gzip",
                    "Authorization": "abc123",
                    "Content-Length": "811",
                    "Host": "httpbin.org",
                    "User-Agent": "Go-http-client/2.0",
                    "X-Amzn-Trace-Id": "Root=1-680f008a-4529fe076da393835ad33165"
                },
                "json": {
                    "assets": [
                        "https://wx2.sinaimg.cn/large/007Raq4zgy1hzm3r1qr85j30zu25o7wh.jpg"
                    ],
                    "avatar": "https://wx1.sinaimg.cn/orj480/007Raq4zly8hzk8uj9qc8j30e80e8mya.jpg",
                    "banner": [
                        "https://wx4.sinaimg.cn/crop.0.0.640.640.640/007Raq4zgy1htmdatc6nyj30u00u0771.jpg"
                    ],
                    "comment_id": null,
                    "comments": null,
                    "created_at": "2025-03-19T03:08:52.835047151Z",
                    "description": "蓝色饭团",
                    "edited": false,
                    "extra": {
                        "is_top": false,
                        "source": "🦈iPhone 14 Pro Max"
                    },
                    "follower": "103.6万",
                    "following": "212",
                    "id": 1276,
                    "mid": "5145866868364764",
                    "name": "七海Nana7mi",
                    "platform": "weibo",
                    "reply": null,
                    "reply_id": null,
                    "source": "发布于 上海",
                    "submitter": "188888131",
                    "text": "太好听噜 大早上给我听醉了[种树] ",
                    "time": "2025-03-19T11:08:50+08:00",
                    "title": "",
                    "type": "blog",
                    "uid": "7198559139",
                    "url": "https://m.weibo.cn/status/PjfcOz63y"
                },
                "origin": "111.117.120.180",
                "url": "https://httpbin.org/post"
            },
            "error": ""
        }
    ]
}
```

### GET /api/logs/*filepath

管理员鉴权成功后，查看日志。

```
/api/logs
```

```
2025-04-16.log
2025-04-17.log
2025-04-18.log
```

### GET /api/root/*filepath

所有者鉴权成功后，查看运行目录。

```
/api/root
```

```
blogs.db
logs/
public/
users.db
```

### POST /api/upload

所有者鉴权成功后，使用 `MultipartForm` 请求体上传文件。上传文件中的 `index.html` 会被重命名并设置成新的主页。

```json
{
    "code": 0,
    "data": "success"
}
```

### GET /api/execute

所有者鉴权成功后，执行命令，请求参数包括：

| 参数名 | 参数类型 | 参数含义                              |
| ------ | -------- | ------------------------------------- |
| cmd    | string   | 命令                                  |
| dir    | string   | 执行命令的路径                        |
| keep   | bool     | 执行行 `Windows` 命令是否用 `/K` 参数 |

```
/api/execute?cmd=ver
```

```json
{
    "code": 0,
    "data": [
        "Microsoft Windows [版本 10.0.26100.3775]"
    ]
}
```

### GET /api/shutdown

所有者鉴权成功后，优雅关机。

```json
{
    "code": 0,
    "data": "人生有梦，各自精彩！"
}
```

