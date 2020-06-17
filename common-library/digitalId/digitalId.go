package digitalId

import (
	"servers/common-library/log"
	"servers/common-library/redisKeyPrefix"

	"github.com/hh8456/go-common/redisObj"
)

const key = "id"

func Gen() {
	// 用 redis scard 方法查询集合中元素的数量
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.DigitalId)
	cnt, e := rds.Scard(key)
	if e != nil {
		log.Errorf("digitalId.Gen - redis.Scard 访问 redis 发生错误: %v\n", e)
		return
	}

	if cnt > 0 {
		log.Warnf("redis 中集合 %s 还有元素, 不需要往该集合中增加元素\n", rds.GetPrefix()+key)
		return
	}

	// 如果没有元素数量了,就用 sadd 发放,向集合中增加一批元素
	ids := []interface{}{}
	for i := 100000; i < 1000000; i++ {
		ids = append(ids, i)
	}
	rds.AddSetMembers(key, ids...)
}

func Get() string {
	// 调用 redis spop 方法,移除并返回及集合中一个随机元素
	rds := redisObj.NewSessionWithPrefix(redisKeyPrefix.DigitalId)
	strId, e := rds.PopSetMember(key)
	if e != nil {
		log.Errorf("digitalId.Gen - redis.spop 访问 redis 发生错误%v\n", e)
		return ""
	}

	return strId
}
