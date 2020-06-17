package utility

import (
	"math/rand"
	"reflect"
	"servers/common-library/proto/commonProto"
	"time"
)

// 获得一堆扑克牌
func GetPokerHeap() []*commonProto.PokerCard {
	pokerCards := make([]*commonProto.PokerCard, 0, 52)
	for color := commonProto.PokerColor_pokerColorBox; color <= commonProto.PokerColor_pokerColorSpades; color++ {
		for num := commonProto.PokerNum_pokerNumA; num <= commonProto.PokerNum_pokerNumK; num++ {
			pokerCard := &commonProto.PokerCard{}
			pokerCard.PokerNum = int32(num)
			pokerCard.PokerColor = int32(color)
			pokerCards = append(pokerCards, pokerCard)
		}
	}

	RandSlice(pokerCards)
	return pokerCards
}

func RandSlice(slice interface{}) {
	rv := reflect.ValueOf(slice)
	if rv.Type().Kind() != reflect.Slice {
		return
	}

	length := rv.Len()
	if length < 2 {
		return
	}

	swap := reflect.Swapper(slice)
	rand.Seed(time.Now().UnixNano())
	for i := length - 1; i >= 0; i-- {
		j := rand.Intn(length)
		swap(i, j)
	}
}
