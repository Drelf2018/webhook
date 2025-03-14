package api

// 自动下载
//
// 开启会极大的占用带宽，建议发送完所有 hook 请求后再下载
var AutoDownload bool

const AutoDownloadKey string = "auto_download"

// 用于 JWT 加密的密钥
var JWTSecretKey []byte

const JWTSecretKeyKey string = "jwt_secret_key"

// 自动另存数据库
var AutoSave bool

const AutoSaveKey string = "auto_save"

// 从配置中读取值，如果不存在则写入
func LoadOrStore[T any](extra map[string]any, key string, value T) (actual T, loaded bool) {
	if extra == nil {
		return
	}
	v, ok := extra[key]
	if ok {
		actual, loaded = v.(T)
		return
	}
	extra[key] = value
	return value, false
}
