package niuniuAlgorithm

import (
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/niuniuProto"
)

// 顺子牛, 该牌型不需要有牛，只要5张牌组成顺子即可；10JQKA也是顺子
func SeqCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		storedCards := niuniuProto.StoredCardSlice{}
		for _, card := range cards {
			storedCards = append(storedCards, card)
		}
		storedCards.Sort()

		b := true
		for i := 0; i < 4; i++ {
			if storedCards[i].PokerNum+1 == storedCards[i+1].PokerNum {
				b = b && true
			} else {
				b = b && false
			}
		}

		if b {
			return true
		}

		// 判断特殊牌型 10JQKA
		// 排序后 ==>   A10JQK
		if storedCards[0].PokerNum == int32(commonProto.PokerNum_pokerNumA) &&
			storedCards[1].PokerNum == int32(commonProto.PokerNum_pokerNum10) &&
			storedCards[2].PokerNum == int32(commonProto.PokerNum_pokerNumJ) &&
			storedCards[3].PokerNum == int32(commonProto.PokerNum_pokerNumQ) &&
			storedCards[4].PokerNum == int32(commonProto.PokerNum_pokerNumK) {
			return true
		}
	}

	return false
}

// 10 J Q K A 是最大的顺子
func isMaxSeq(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		storedCards := niuniuProto.StoredCardSlice{}
		for _, card := range cards {
			storedCards = append(storedCards, card)
		}
		storedCards.Sort()

		b1 := storedCards[0].PokerNum == int32(commonProto.PokerNum_pokerNumA)
		b2 := storedCards[1].PokerNum == int32(commonProto.PokerNum_pokerNum10)
		b3 := storedCards[2].PokerNum == int32(commonProto.PokerNum_pokerNumJ)
		b4 := storedCards[3].PokerNum == int32(commonProto.PokerNum_pokerNumQ)
		b5 := storedCards[4].PokerNum == int32(commonProto.PokerNum_pokerNumK)

		return b1 && b2 && b3 && b4 && b5
	}

	return false
}
