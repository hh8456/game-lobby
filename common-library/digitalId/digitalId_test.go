package digitalId

import (
	"testing"

	"github.com/hh8456/go-common/redisObj"
)

func TestGen(t *testing.T) {
	redisObj.Init("192.168.0.155:6379", "")
	Gen()
	println(Get())
}
