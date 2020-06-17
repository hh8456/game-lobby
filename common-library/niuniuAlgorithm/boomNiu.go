package niuniuAlgorithm

import (
	"servers/common-library/proto/commonProto"
)

// 炸弹牛, 5张牌里，有4张牌点数相同即可
func BoomCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		m := map[int32]int32{} // 点数 - 数量
		for _, card := range cards {
			m[card.PokerNum]++
		}

		for _, num := range m {
			if num == 4 {
				return true
			}
		}
	}

	return false
}
