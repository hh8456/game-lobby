package niuniuAlgorithm

import (
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/niuniuProto"
)

// 葫芦牛, 该牌型不需要有牛，只要三张+对子的组合即可
func HuluCardType(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		m := map[int32]int32{} // 牌面值 - 数量
		for _, card := range cards {
			m[card.PokerNum]++
		}

		for _, cnt := range m {
			if false == ((cnt == 3) || (cnt == 2)) {
				return false
			}
		}

		return true
	}

	return false
}

// AAA 是最大的三张
func isMaxHulu(cards []*commonProto.PokerCard) bool {
	if len(cards) == 5 {
		storedCards := niuniuProto.StoredCardSlice{}
		for _, card := range cards {
			storedCards = append(storedCards, card)
		}
		storedCards.Sort()

		b1 := storedCards[0].PokerNum == int32(commonProto.PokerNum_pokerNumA)
		b2 := storedCards[1].PokerNum == int32(commonProto.PokerNum_pokerNumA)
		b3 := storedCards[2].PokerNum == int32(commonProto.PokerNum_pokerNumA)

		b4 := (storedCards[3].PokerNum == storedCards[4].PokerNum) &&
			(storedCards[3].PokerNum != storedCards[0].PokerNum)

		return b1 && b2 && b3 && b4
	}

	return false
}
