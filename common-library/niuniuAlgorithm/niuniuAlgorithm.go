package niuniuAlgorithm

import (
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/niuniuProto"
	"sort"
)

type storedInt32Slice []int32

func (s storedInt32Slice) Len() int           { return len(s) }
func (s storedInt32Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s storedInt32Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s storedInt32Slice) Sort()              { sort.Sort(s) }

// 牛牛特殊牌型判断; 返回最大的牌型
func SpecialCardType(cards []*commonProto.PokerCard) niuniuProto.NiuniuCardType {
	// 同花顺 ＞ 五小牛＞ 炸弹牛＞ 葫芦牛＞ 同花牛＞ 无花牛＞ 顺子牛
	if StraightFlushCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_straightFlush
	}

	if FiveCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_five
	}

	if BoomCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_boom
	}

	if HuluCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_hulu
	}

	if SameFlowerCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_sameFlower
	}

	if FiveFlowerCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_fiveFlower
	}

	if SeqCardType(cards) {
		return niuniuProto.NiuniuCardType_niuniuCardType_seq
	}

	return niuniuProto.NiuniuCardType_niuniuCardType_0
}

func normalCardType(total int32, card1, card2, card3 *commonProto.PokerCard, mapCardTypes map[int32][]*commonProto.PokerCard) {
	sum := int32(0)
	if card1.PokerNum < 10 {
		sum += card1.PokerNum
	} else {
		sum += 10
	}

	if card2.PokerNum < 10 {
		sum += card2.PokerNum
	} else {
		sum += 10
	}

	if card3.PokerNum < 10 {
		sum += card3.PokerNum
	} else {
		sum += 10
	}

	if sum%10 == 0 {
		cardType := (total - sum) % 10
		cardComb := []*commonProto.PokerCard{card1, card2, card3}
		if cardType == 0 {
			mapCardTypes[int32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = cardComb // 牛牛
		} else {
			mapCardTypes[cardType] = cardComb // 牛1 - 牛9
		}
	}
}

func NormalCardType(cards []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard) {
	if len(cards) == 5 {
		total := int32(0)
		for i := 0; i < 5; i++ {
			if cards[i].PokerNum > 10 {
				total += 10
			} else {
				total += cards[i].PokerNum
			}
		}

		//  穷举各种牌型组合, 共 10 种, 为了支持癞子算法

		// mapCardTypes 键值对: 牌型( 无牛 - 牛牛 ) - 构成此牌型的组合
		mapCardTypes := map[int32][]*commonProto.PokerCard{}
		// 0, 1, 2 => 3, 4
		normalCardType(total, cards[0], cards[1], cards[2], mapCardTypes)
		// 0, 1, 3 => 2, 4
		normalCardType(total, cards[0], cards[1], cards[3], mapCardTypes)
		// 0, 1, 4 => 2, 3
		normalCardType(total, cards[0], cards[1], cards[4], mapCardTypes)
		// 0, 2, 3 => 1, 4
		normalCardType(total, cards[0], cards[2], cards[3], mapCardTypes)
		// 0, 2, 4 => 1, 3
		normalCardType(total, cards[0], cards[2], cards[4], mapCardTypes)
		// 0, 3, 4 => 1, 2
		normalCardType(total, cards[0], cards[3], cards[4], mapCardTypes)
		// 1, 2, 3 => 0, 4
		normalCardType(total, cards[1], cards[2], cards[3], mapCardTypes)
		// 1, 2, 4 => 0, 3
		normalCardType(total, cards[1], cards[2], cards[4], mapCardTypes)
		// 1, 3, 4 => 0, 2
		normalCardType(total, cards[1], cards[3], cards[4], mapCardTypes)
		// 2, 3, 4 => 0, 1
		normalCardType(total, cards[2], cards[3], cards[4], mapCardTypes)

		cardTypes := storedInt32Slice{}
		for cardType, _ := range mapCardTypes {
			cardTypes = append(cardTypes, cardType)
		}
		cardTypes.Sort()

		cardType := int32(0)
		if len(cardTypes) == 0 {
			// 无牛
			return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}
		} else {
			if len(cardTypes) > 1 {
				log.Debugf("牛牛, 出现多种普通牌型, cards: %v, cardTypes: %v", cards, cardTypes)
			}
			//cardType 的值是 [1, 10]
			cardType = cardTypes[len(cardTypes)-1]
			return niuniuProto.NiuniuCardType(cardType), mapCardTypes[cardType]
		}
	}

	return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}
}

// 牛牛普通牌型判断; 返回最大的牌型  //XXX 20.6.1 日注释,系统稳定后删除. jason
//func NormalCardType2(cards []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard) {

//if len(cards) == 5 {
//storedCards := niuniuProto.StoredCardSlice{}
//for _, card := range cards {
//storedCards = append(storedCards, card)
//}
//storedCards.Sort()

//total := int32(0)
//for i := 0; i < 5; i++ {
//if storedCards[i].PokerNum > 10 {
//total += 10
//} else {
//total += storedCards[i].PokerNum
//}
//}

//mapCardTypes := map[int32][]*commonProto.PokerCard{}
//// 各种牌型组合
//for a := 0; a < 3; a++ {
//for b := a + 1; b < 4; b++ {
//for c := b + 1; c < 5; c++ {
//sum := int32(0)
//if storedCards[a].PokerNum > 10 {
//sum += 10
//} else {
//sum += storedCards[a].PokerNum
//}

//if storedCards[b].PokerNum > 10 {
//sum += 10
//} else {
//sum += storedCards[b].PokerNum
//}

//if storedCards[c].PokerNum > 10 {
//sum += 10
//} else {
//sum += storedCards[c].PokerNum
//}

//// 三张牌点数之和 等于 10 或者是 10 的倍数, 才能算有牛
//if sum%10 == 0 {
//cardType := (total - sum) % 10
//cardComb := []*commonProto.PokerCard{storedCards[a], storedCards[b], storedCards[c]}
//if cardType == 0 {
//mapCardTypes[int32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = cardComb // 牛牛
//} else {
//mapCardTypes[cardType] = cardComb // 牛1 - 牛9
//}
//}
//}
//}
//}

//cardTypes := storedInt32Slice{}
//for cardType, _ := range mapCardTypes {
//cardTypes = append(cardTypes, cardType)
//}
//cardTypes.Sort()

//cardType := int32(0)
//if len(cardTypes) == 0 {
//// 无牛
//return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}
//} else {
//if len(cardTypes) > 1 {
//log.Debugf("牛牛, 出现多种普通牌型, cards: %v, cardTypes: %v", cards, cardTypes)
//}
////cardType 的值是 [1, 10]
//cardType = cardTypes[len(cardTypes)-1]
//return niuniuProto.NiuniuCardType(cardType), mapCardTypes[cardType]
//}
//}

//return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}
//}

// 计算牌型, 优先取特殊牌型,其次取普通牌型
// 返回值的第二个参数,是指牌型组合
func CalCardType(cards []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard) {
	specCardType := SpecialCardType(cards)
	if specCardType != niuniuProto.NiuniuCardType_niuniuCardType_0 {
		return specCardType, cards
	}

	return NormalCardType(cards)
}

func CardTypeString(cardType niuniuProto.NiuniuCardType) string {
	switch cardType {
	case niuniuProto.NiuniuCardType_niuniuCardType_0:
		return "无牛"

	case niuniuProto.NiuniuCardType_niuniuCardType_1:
		return "牛一"

	case niuniuProto.NiuniuCardType_niuniuCardType_2:
		return "牛二"

	case niuniuProto.NiuniuCardType_niuniuCardType_3:
		return "牛三"

	case niuniuProto.NiuniuCardType_niuniuCardType_4:
		return "牛四"

	case niuniuProto.NiuniuCardType_niuniuCardType_5:
		return "牛五"

	case niuniuProto.NiuniuCardType_niuniuCardType_6:
		return "牛六"

	case niuniuProto.NiuniuCardType_niuniuCardType_7:
		return "牛七"

	case niuniuProto.NiuniuCardType_niuniuCardType_8:
		return "牛八"

	case niuniuProto.NiuniuCardType_niuniuCardType_9:
		return "牛九"

	case niuniuProto.NiuniuCardType_niuniuCardType_niu:
		return "牛牛"

	case niuniuProto.NiuniuCardType_niuniuCardType_seq:
		return "顺子牛"

	// 五花牛, 有牛，且5张牌均是 10/J/Q/K 中的一种
	case niuniuProto.NiuniuCardType_niuniuCardType_fiveFlower:
		return "五花牛"

	// 同花牛, 该牌型不需要有牛，只要5张牌为同一花色即可
	case niuniuProto.NiuniuCardType_niuniuCardType_sameFlower:
		return "同花牛"

	// 葫芦牛, 该牌型不需要有牛，只要三张+对子的组合即可
	case niuniuProto.NiuniuCardType_niuniuCardType_hulu:
		return "葫芦牛"

	// 炸弹牛, 5张牌里，有4张牌点数相同即可
	case niuniuProto.NiuniuCardType_niuniuCardType_boom:
		return "炸弹牛"

	// 五小牛, 5张牌的点数加起来不超过10（不含10）
	case niuniuProto.NiuniuCardType_niuniuCardType_five:
		return "五小牛"

	// 同花顺, 5张花色相同的顺子；10JQKA也是顺子
	case niuniuProto.NiuniuCardType_niuniuCardType_straightFlush:
		return "同花顺"
	}

	return "无牛"
}

// cards_1 能赢 cards_2 就返回 true, 否则返回 false
func Compare(cards_1, cards_2 []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard, niuniuProto.NiuniuCardType, []*commonProto.PokerCard, bool) {
	cardType1, cardComb1 := CalCardType(cards_1)
	cardType2, cardComb2 := CalCardType(cards_2)
	if cardType1 > cardType2 {
		return cardType1, cardComb1, cardType2, cardComb2, true
	}

	if cardType1 < cardType2 {
		return cardType1, cardComb1, cardType2, cardComb2, false
	}

	// 牌型一样,就用点数和花色比
	// cardType1 == cardType2
	switch cardType1 {
	// 无牛, 牛1,.....牛9, 牛牛, 五花牛, 五小牛, 同花牛
	case niuniuProto.NiuniuCardType_niuniuCardType_0,
		niuniuProto.NiuniuCardType_niuniuCardType_1,
		niuniuProto.NiuniuCardType_niuniuCardType_2,
		niuniuProto.NiuniuCardType_niuniuCardType_3,
		niuniuProto.NiuniuCardType_niuniuCardType_4,
		niuniuProto.NiuniuCardType_niuniuCardType_5,
		niuniuProto.NiuniuCardType_niuniuCardType_6,
		niuniuProto.NiuniuCardType_niuniuCardType_7,
		niuniuProto.NiuniuCardType_niuniuCardType_8,
		niuniuProto.NiuniuCardType_niuniuCardType_9,
		niuniuProto.NiuniuCardType_niuniuCardType_niu,
		niuniuProto.NiuniuCardType_niuniuCardType_fiveFlower,
		niuniuProto.NiuniuCardType_niuniuCardType_five,
		niuniuProto.NiuniuCardType_niuniuCardType_sameFlower:

		card1, card2 := getMaxCard(cards_1), getMaxCard(cards_2)
		return cardType1, cardComb1, cardType2, cardComb2, compare(card1, card2)

	// 葫芦牛
	case niuniuProto.NiuniuCardType_niuniuCardType_hulu:
		return cardType1, cardComb1, cardType2, cardComb2, compareHulu(cards_1, cards_2)

	// 炸弹牛, AAAA 是最大的炸弹
	case niuniuProto.NiuniuCardType_niuniuCardType_boom:
		return cardType1, cardComb1, cardType2, cardComb2, compareBoom(cards_1, cards_2)

	// 顺子牛, 该牌型不需要有牛，只要5张牌组成顺子即可；10JQKA也是顺子
	case niuniuProto.NiuniuCardType_niuniuCardType_seq,
		// 同花顺, 5张花色相同的顺子；10JQKA也是顺子
		niuniuProto.NiuniuCardType_niuniuCardType_straightFlush:
		return cardType1, cardComb1, cardType2, cardComb2, compareSeq(cards_1, cards_2)

	}

	return cardType1, cardComb1, cardType2, cardComb2, false
}

// 两个葫芦比大小, AAA 是最大的
func compareHulu(cards_1, cards_2 []*commonProto.PokerCard) bool {
	// 两边都是最大的三张,就用 A 的花色来比较
	if isMaxHulu(cards_1) && isMaxHulu(cards_2) {
		color1 := commonProto.PokerColor_pokerColorBox
		color2 := commonProto.PokerColor_pokerColorBox

		for _, card := range cards_1 {
			if card.PokerNum == int32(commonProto.PokerNum_pokerNumA) {
				color1 = commonProto.PokerColor(card.PokerColor)
				break
			}
		}

		for _, card := range cards_2 {
			if card.PokerNum == int32(commonProto.PokerNum_pokerNumA) {
				color2 = commonProto.PokerColor(card.PokerColor)
				break
			}
		}

		if color1 > color2 {
			return true
		}

		return false
	}

	if isMaxHulu(cards_1) {
		return true
	}

	if isMaxHulu(cards_2) {
		return false
	}

	card1, card2 := getMaxCard(cards_1), getMaxCard(cards_2)
	return compare(card1, card2)
}

// 2个顺子比大小
func compareSeq(cards_1, cards_2 []*commonProto.PokerCard) bool {
	// 两边都是最大的顺子,就用 A 的花色来比较
	if isMaxSeq(cards_1) && isMaxSeq(cards_2) {
		color1 := commonProto.PokerColor_pokerColorBox
		color2 := commonProto.PokerColor_pokerColorBox

		for _, card := range cards_1 {
			if card.PokerNum == int32(commonProto.PokerNum_pokerNumA) {
				color1 = commonProto.PokerColor(card.PokerColor)
				break
			}
		}

		for _, card := range cards_2 {
			if card.PokerNum == int32(commonProto.PokerNum_pokerNumA) {
				color2 = commonProto.PokerColor(card.PokerColor)
				break
			}
		}

		if color1 > color2 {
			return true
		}

		return false
	}

	if isMaxSeq(cards_1) {
		return true
	}

	if isMaxSeq(cards_2) {
		return false
	}

	card1, card2 := getMaxCard(cards_1), getMaxCard(cards_2)
	return compare(card1, card2)
}

// 2个炸弹牛比大小
func compareBoom(cards_1, cards_2 []*commonProto.PokerCard) bool {
	mapBoom1 := map[int32]int32{}
	for _, card := range cards_1 {
		mapBoom1[card.PokerNum]++
	}

	num1 := int32(0)
	for num, cnt := range mapBoom1 {
		if cnt == 4 {
			num1 = num
		}

	}

	mapBoom2 := map[int32]int32{}
	for _, card := range cards_2 {
		mapBoom2[card.PokerNum]++
	}

	num2 := int32(0)
	for num, cnt := range mapBoom2 {
		if cnt == 4 {
			num2 = num
		}
	}

	// AAAA 是最大的炸弹
	if num1 == int32(commonProto.PokerNum_pokerNumA) {
		return true
	}

	if num2 == int32(commonProto.PokerNum_pokerNumA) {
		return false
	}

	if num1 > num2 {
		return true
	}

	return false
}

// card1 能赢 card2 就返回 true, 否则返回 false
func compare(card1, card2 *commonProto.PokerCard) bool {
	if card1.PokerNum > card2.PokerNum {
		return true
	}

	if card1.PokerNum < card2.PokerNum {
		return false
	}

	// card1.PokerNum == card2.PokerNum
	if card1.PokerColor > card2.PokerColor {
		return true
	}

	return false
}

// 按最大点数取牌,如果有多个相同点数的牌,就按最大花色取
func getMaxCard(cards []*commonProto.PokerCard) *commonProto.PokerCard {
	if len(cards) == 5 {
		storedCards := niuniuProto.StoredCardSlice{}
		for _, card := range cards {
			storedCards = append(storedCards, card)
		}
		storedCards.Sort()

		colorValues := storedInt32Slice{}
		maxNum := storedCards[len(storedCards)-1].PokerNum
		for _, card := range cards {
			if card.PokerNum == maxNum {
				colorValues = append(colorValues, card.PokerColor)
			}
		}
		colorValues.Sort()
		maxColor := colorValues[len(colorValues)-1]

		return &commonProto.PokerCard{PokerNum: maxNum,
			PokerColor: maxColor}
	}

	// 默认取最小点数,最小花色
	return &commonProto.PokerCard{PokerNum: int32(commonProto.PokerNum_pokerNumA),
		PokerColor: int32(commonProto.PokerColor_pokerColorBox)}
}
