package niuniuAlgorithm

import (
	"servers/common-library/proto/commonProto"
)

// 五花牛, 有牛，且5张牌均是 10/J/Q/K 中的一种
func FiveFlowerCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		for _, card := range cards {
			if card.PokerNum < 10 {
				return false
			}
		}

		return true
	}

	return false
}
