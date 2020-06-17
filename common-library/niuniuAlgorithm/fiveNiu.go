package niuniuAlgorithm

import "servers/common-library/proto/commonProto"

// 五小牛, 5张牌的点数加起来不超过10（不含10）
func FiveCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		cnt := int32(0)
		for _, card := range cards {
			cnt += card.PokerNum
		}

		if cnt < 10 {
			return true
		}
	}

	return false
}
