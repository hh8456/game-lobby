package niuniuAlgorithm

import "servers/common-library/proto/commonProto"

// 同花牛, 该牌型不需要有牛，只要5张牌为同一花色即可
func SameFlowerCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		color := cards[0].PokerColor
		for i := 1; i < 5; i++ {
			if cards[i].PokerColor != color {
				return false
			}
		}

		return true
	}

	return false
}
