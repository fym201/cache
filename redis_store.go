package cache

import (
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"log"

	"reflect"
	"strconv"
)

// redis store
type RedisStore struct {
	uri    string //redis服务器连接地址
	domain string //作用域
	pool   *redis.Pool
}

// 重写生成连接池方法
func newPool(uri string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", uri)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func NewRedisStore(uri string, domain string) *RedisStore {
	store := &RedisStore{uri: uri, pool: newPool(uri), domain: domain}
	return store
}

func (s *RedisStore) Expire(key string, sec int64) (err error) {
	key = s.domain + key
	ReTry:
	conn := s.pool.Get()
	defer conn.Close()

	_, err = conn.Do("EXPIRE", key, sec);
	if err != nil {
		log.Println("EXPIRE ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
	}
	return nil
}

func (s *RedisStore) TTL(key string) (int64) {
	key = s.domain + key
	ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	rep, err := conn.Do("TTL", key);
	if err != nil {
		log.Println("TTL ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
	}

	return rep.(int64)
}

func (s *RedisStore) Put(key string, value interface{}) (err error) {
	key = s.domain + key
	var data interface{}
	val := reflect.TypeOf(value)

	for  {
		kd := val.Kind()
		if kd == reflect.Struct || kd ==  reflect.Map || kd == reflect.Slice {
			data, err = bson.Marshal(value)
			if err != nil {
				log.Fatal(err)
				return err
			}
			break
		} else if kd == reflect.Ptr {
			val = val.Elem()
			continue
		} else {
			data = value
			break
		}
	}


	ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	if _, err = conn.Do("SET", key, data); err != nil {
		log.Println("Put ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
	}
	return nil
}

func (s *RedisStore) PutWithExpire(key string, value interface{}, sec int64) (err error) {
	err = s.Put(key, value)
	if (err != nil) {
		return err
	}
	err = s.Expire(key, sec)
	if err != nil {
		s.Del(key)
	}
	return err
}

func (s *RedisStore) Get(key string, out interface{}) error {
	key = s.domain + key

	ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	ret, err := conn.Do("GET", key)
	if err != nil {
		log.Println("GetObject ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
	}

	if ret != nil {
		err = bson.Unmarshal(ret.([]byte), out)
		return nil
	}
	return ErrNotExist
}

func (s *RedisStore) GetString(key string) (string, error) {
	key = s.domain + key
	ReTry:
	conn := s.pool.Get()
	defer conn.Close()
	ret, err := conn.Do("GET", key)
	if err != nil {
		log.Println("GetString ERROR:", err.Error())
		if err == nil {
			goto ReTry
		}
		return "", err
	}

	if ret == nil {
		return "", ErrNotExist
	}
	return string(ret.([]byte)), nil
}

func (s *RedisStore) GetMustString(key string) string {
	str, _ := s.GetString(key)
	return str
}

func (s *RedisStore) GetInt(key string) (int, error) {
	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(ret)
}

func (s *RedisStore) GetMustInt(key string) int {
	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}
	num, _ := strconv.Atoi(ret)
	return num
}

func (s *RedisStore) GetInt64(key string) (int64, error) {
	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(ret, 10, 64)
}

func (s *RedisStore) GetMustInt64(key string) int64 {
	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}
	num, _ := strconv.ParseInt(ret, 10, 64)
	return num
}

func (s *RedisStore) GetFloat32(key string) (float32, error) {
	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseFloat(ret, 32)
	return float32(num), err
}

func (s *RedisStore) GetMustFloat32(key string) float32 {
	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}
	num, _ := strconv.ParseFloat(ret, 64)
	return float32(num)
}

func (s *RedisStore) GetFloat64(key string) (float64, error) {
	ret, err := s.GetString(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(ret, 64)
}

func (s *RedisStore) GetMustFloat64(key string) float64 {
	ret, err := s.GetString(key)
	if err != nil {
		return 0
	}
	num, _ := strconv.ParseFloat(ret, 64)
	return num
}

func (s *RedisStore) Del(key string) error {
	key = s.domain + key
	conn := s.pool.Get()
	defer conn.Close()
	_, e := conn.Do("DEL", key)
	return e
}

func (s *RedisStore) Destroy() {

}