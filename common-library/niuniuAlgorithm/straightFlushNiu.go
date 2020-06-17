package niuniuAlgorithm

import "servers/common-library/proto/commonProto"

//同花顺, 5张花色相同的顺子；10JQKA也是顺子

func StraightFlushCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		// 判断花色相同
		color := cards[0].PokerColor
		for i := 1; i < 5; i++ {
			if cards[i].PokerColor != color {
				return false
			}
		}

		return SeqCardType(cards)
	}

	return false
}
