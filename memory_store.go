package cache

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"qiniupkg.com/x/errors.v7"
)

// memory store
type MemoryStore struct {
	mutex     sync.RWMutex
	store     map[string]interface{}
	expire    map[string]int64
	destroyed bool
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{store: make(map[string]interface{}), expire: make(map[string]int64)}
	go s.expireLoop()
	return s
}

func (s *MemoryStore) expireLoop() {
	for {
		if s.destroyed {
			return
		}
		select {
		case <-time.After(time.Millisecond * 10):
			s.mutex.Lock()
			now := time.Now().UnixNano()
			for k, v := range s.expire {
				if v-now <= 0 {
					delete(s.expire, k)
					delete(s.store, k)
				}
			}
			s.mutex.Unlock()
		}
	}
}

func (s *MemoryStore) Expire(key string, sec int64) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.expire[key] = (sec * int64(time.Second)) + time.Now().UnixNano()
	return nil
}

func (s *MemoryStore) TTL(key string) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, ok := s.expire[key]; ok {
		ttl := int64((v - time.Now().UnixNano()) / int64(time.Second))
		if ttl < 0 {
			return 0
		}
		return ttl
	}
	return -1
}

func (s *MemoryStore) Put(key string, value interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
	return nil
}

func (s *MemoryStore) PutWithExpire(key string, value interface{}, sec int64) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
	s.expire[key] = (sec * int64(time.Second)) + time.Now().UnixNano()
	return nil
}

func (s *MemoryStore) Get(key string, out interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		vl := reflect.ValueOf(out).Elem()
		if !vl.CanSet() {
			return errors.New("out cannot set")
		}
		vl.Set(reflect.ValueOf(v).Elem())
		return nil
	}
	return ErrNotExist
}

func (s *MemoryStore) GetString(key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		switch v.(type) {
		case string:
			return string(v.(string)), nil
		case []byte:
			return string(v.([]byte)), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "",ErrNotExist
}

func (s *MemoryStore) GetMustString(key string) string {
	str, _ := s.GetString(key)
	return str
}

func (s *MemoryStore) GetInt(key string) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		n, err := toInt(v)
		return int(n), err
	}
	return 0, ErrNotExist
}

func (s *MemoryStore) GetMustInt(key string) int {
	ret, _ := s.GetInt(key)
	return ret
}

func (s *MemoryStore) GetInt64(key string) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		n, err := toInt(v)
		return n, err
	}
	return 0, ErrNotExist
}

func (s *MemoryStore) GetMustInt64(key string) int64 {
	ret, _ := s.GetInt64(key)
	return ret
}

func (s *MemoryStore) GetFloat32(key string) (float32, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		n, err := toInt(v)
		return float32(n), err
	}
	return 0, ErrNotExist
}

func (s *MemoryStore) GetMustFloat32(key string) float32 {
	ret, _ := s.GetFloat32(key)
	return ret
}

func (s *MemoryStore) GetFloat64(key string) (float64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		n, err := toFloat(v)
		return n, err
	}
	return 0, ErrNotExist
}

func (s *MemoryStore) GetMustFloat64(key string) float64 {
	ret, _ := s.GetFloat64(key)
	return ret
}

func (s *MemoryStore) Del(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
	return nil
}

func (s *MemoryStore) Destroy() {
	s.destroyed = true
}

func toInt(value interface{}) (int64, error) {
	var ret int64 = 0
	switch v := value.(type) {
	case bool:
		if true {
			ret = 1
		}
	case float32:
		ret = int64(v)
	case float64:
		ret = int64(v)
	case int:
		ret = int64(v)
	case int8:
		ret = int64(v)
	case int16:
		ret = int64(v)
	case int32:
		ret = int64(v)
	case int64:
		ret = v
	case uint:
		ret = int64(v)
	case uint8:
		ret = int64(v)
	case uint16:
		ret = int64(v)
	case uint32:
		ret = int64(v)
	case uint64:
		ret = int64(v)
	case string:
		ret, _ = strconv.ParseInt(v, 10, 64)
	default:
		return 0, errors.New("Can not convert to int")
	}
	return ret, nil
}

func toFloat(value interface{}) (float64, error) {
	var ret float64 = 0
	switch v := value.(type) {
	case bool:
		if true {
			ret = 1.0
		}
	case float32:
		ret = float64(v)
	case float64:
		ret = v
	case int:
		ret = float64(v)
	case int8:
		ret = float64(v)
	case int16:
		ret = float64(v)
	case int32:
		ret = float64(v)
	case int64:
		ret = float64(v)
	case uint:
		ret = float64(v)
	case uint8:
		ret = float64(v)
	case uint16:
		ret = float64(v)
	case uint32:
		ret = float64(v)
	case uint64:
		ret = float64(v)
	case string:
		ret, _ = strconv.ParseFloat(v, 64)
	default:
		return 0, errors.New("Can not convert to float64")
	}
	return ret, nil
}
