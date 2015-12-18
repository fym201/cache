package cache
import "errors"

var (
	ErrUnSupportedType error = errors.New("Unsupported type error")
	ErrNotExist error = errors.New("Not exist error")
)

type CacheStore interface  {
	Expire(key string, sec int64) (err error)
	TTL(key string) (int64)
	Put(key string, value interface{}) (err error)
	PutWithExpire(key string, value interface{}, sec int64) (err error)
	Get(key string, out interface{}) error
	GetString(key string) (string, error)
	GetMustString(key string) string
	GetInt(key string) (int, error)
	GetMustInt(key string) int
	GetInt64(key string) (int64, error)
	GetMustInt64(key string) int64
	GetFloat32(key string) (float32, error)
	GetMustFloat32(key string) float32
	GetFloat64(key string) (float64, error)
	GetMustFloat64(key string) float64
	Del(key string)
	Destroy()
}