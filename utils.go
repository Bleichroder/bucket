package bucket

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// Bucket 超时桶 元素放到桶里,超时的桶倒掉
type bucket struct {
	Tag string
	// bigMap  *sync.Map
	bigMap  sync.Map
	cap     int
	buckets []*list.List
	timeout int
}

// ToEle 插入结构必须符合此类 kv特征.
type ToEle interface {
	GetID() string
}

type element struct {
	val interface{}
	L   *list.List
	Ele *list.Element
}

type toEleString string

func (t toEleString) GetID() string {
	return string(t)
}

/*
单位（秒）
timeout 缓存时长
step 步长
容量 = 缓存时长／步长
*/
func NewBucket(timeout int, step int) (buc *bucket) {
	cap := timeout / step
	buc = new(bucket)
	buc.timeout = timeout
	buc.cap = cap

	for i := 0; i < cap; i++ {
		buc.buckets = append(buc.buckets, list.New())
	}
	go func() {
		timer := time.NewTimer(time.Second * time.Duration(step))
		for range timer.C {
			timer.Reset(time.Second * time.Duration(step))
			tmp := buc.buckets[buc.cap-1]
			for i := buc.cap - 1; i > 0; i-- {
				buc.buckets[i] = buc.buckets[i-1]
			}
			buc.buckets[0] = list.New()
			for i := tmp.Front(); i != nil; i = i.Next() {
				e, ok := i.Value.(string)
				if !ok {
					fmt.Println("no toele")
				}

				buc.bigMap.Delete(e)
			}
		}
	}()

	return
}

//将数据放入桶中
func (buc *bucket) Push(key string, val interface{}) {
	// _, ok := b.bigMap.Load(e.GetID())
	old, ok := buc.bigMap.Load(key)
	if ok {
		e := old.(element)
		e.L.Remove(e.Ele)
	}

	l := buc.buckets[0]
	ele := l.PushFront(key)
	//3.做缓存
	buc.bigMap.Store(key, element{val: val, L: l, Ele: ele})
}

// Get 根据id 获取缓存中的数据
func (buc *bucket) Get(id string) (value interface{}, ok bool) {
	value, ok = buc.bigMap.Load(id)
	if ok {
		e := value.(element)
		value = e.val
	}
	return
}
