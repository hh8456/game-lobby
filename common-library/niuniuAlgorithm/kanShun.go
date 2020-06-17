package niuniuAlgorithm

import (
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/niuniuProto"
)

// 是否坎斗, 有3个点数相同的就算有牛. eg: 555 就是有牛
// 坎斗属于普通牌型,所以这里不用计算炸弹牛(5 张牌里，有 4 张牌点数相同时就是炸弹牛)
// 返回值: 牛牛牌型, 扑克牌组合, 是否能够坎斗
func Kan(cards []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard, bool) {
	if len(cards) == 5 {
		// 扑克牌点数, 对应的具体牌
		mapCardTypes := map[int32][]*commonProto.PokerCard{}
		// 扑克牌点数, 数量
		m := map[int32]int{}
		for _, card := range cards {
			// A, 2, J, Q, K; ==> 坎斗 JQK 1+2，顺斗QKA 10+2；JQK 1+2
			pokerNum := int32(0)
			if card.PokerNum > 10 {
				pokerNum = 10
			} else {
				pokerNum = card.PokerNum
			}
			m[pokerNum]++
			mapCardTypes[pokerNum] = append(mapCardTypes[card.PokerNum], card)
		}

		var cardsRet []*commonProto.PokerCard
		sum := int32(0)
		// 5 张牌里，有 3(或以上) 张牌点数相同时
		for pokerNum, num := range m {
			if num > 2 {
				cardsRet = mapCardTypes[pokerNum]
			} else {
				if pokerNum < 10 {
					sum += pokerNum
				} else {
					sum += 10
				}
			}
		}

		cardType := sum % 10
		if cardsRet != nil { // 能够进行坎斗
			if cardType == 0 {
				return niuniuProto.NiuniuCardType_niuniuCardType_niu, cardsRet, true
			} else {
				return niuniuProto.NiuniuCardType(cardType), cardsRet, true
			}
		} else {
			// 不能进行坎斗
			return niuniuProto.NiuniuCardType_niuniuCardType_niu, []*commonProto.PokerCard{}, false
		}
	}

	return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}, false
}

// 是否顺斗
// 返回值: 牛牛牌型, 扑克牌组合, 是否能够顺斗
func Shun(cards []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard, bool) {
	if len(cards) == 5 {
		storedCards := niuniuProto.StoredCardSlice{}
		for _, card := range cards {
			storedCards = append(storedCards, card)
		}
		storedCards.Sort()

		// mapCardTypes 键值对: 牌型( 无牛 - 牛牛 ) - 构成此牌型的组合
		mapCardTypes := map[int32][]*commonProto.PokerCard{}
		// 判断 [0,1,2]
		if storedCards[0].PokerNum+1 == storedCards[1].PokerNum &&
			storedCards[0].PokerNum+2 == storedCards[2].PokerNum {

			cardsRet := []*commonProto.PokerCard{storedCards[0],
				storedCards[1], storedCards[2]}

			sum := int32(0)
			if storedCards[3].PokerNum < 10 {
				sum += storedCards[3].PokerNum
			} else {
				sum += 10
			}

			if storedCards[4].PokerNum < 10 {
				sum += storedCards[4].PokerNum
			} else {
				sum += 10
			}

			cardType := sum % 10
			if cardType == 0 {
				mapCardTypes[int32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = cardsRet // 牛牛
			} else {
				mapCardTypes[cardType] = cardsRet // 牛1 - 牛9
			}

		}

		// 判断 [1,2,3]
		if storedCards[1].PokerNum+1 == storedCards[2].PokerNum &&
			storedCards[1].PokerNum+2 == storedCards[3].PokerNum {
			cardsRet := []*commonProto.PokerCard{storedCards[1],
				storedCards[2], storedCards[3]}

			sum := int32(0)
			if storedCards[0].PokerNum < 10 {
				sum += storedCards[0].PokerNum
			} else {
				sum += 10
			}

			if storedCards[4].PokerNum < 10 {
				sum += storedCards[4].PokerNum
			} else {
				sum += 10
			}

			cardType := sum % 10
			if cardType == 0 {
				mapCardTypes[int32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = cardsRet // 牛牛
			} else {
				mapCardTypes[cardType] = cardsRet // 牛1 - 牛9
			}
		}

		// 判断 [2,3,4]
		if storedCards[2].PokerNum+1 == storedCards[3].PokerNum &&
			storedCards[2].PokerNum+2 == storedCards[4].PokerNum {
			cardsRet := []*commonProto.PokerCard{storedCards[2],
				storedCards[3], storedCards[4]}

			sum := int32(0)
			if storedCards[0].PokerNum < 10 {
				sum += storedCards[0].PokerNum
			} else {
				sum += 10
			}

			if storedCards[1].PokerNum < 10 {
				sum += storedCards[1].PokerNum
			} else {
				sum += 10
			}

			cardType := sum % 10
			if cardType == 0 {
				mapCardTypes[int32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = cardsRet // 牛牛
			} else {
				mapCardTypes[cardType] = cardsRet // 牛1 - 牛9
			}
		}

		// 判断特殊组合: QKA也算有牛
		if storedCards[3].PokerNum == int32(commonProto.PokerNum_pokerNumQ) &&
			storedCards[4].PokerNum == int32(commonProto.PokerNum_pokerNumK) &&
			storedCards[0].PokerNum == int32(commonProto.PokerNum_pokerNumA) {
			cardsRet := []*commonProto.PokerCard{storedCards[0],
				storedCards[3], storedCards[4]}
			sum := int32(0)
			if storedCards[1].PokerNum < 10 {
				sum += storedCards[1].PokerNum
			} else {
				sum += 10
			}

			if storedCards[2].PokerNum < 10 {
				sum += storedCards[2].PokerNum
			} else {
				sum += 10
			}

			cardType := sum % 10
			if cardType == 0 {
				mapCardTypes[int32(niuniuProto.NiuniuCardType_niuniuCardType_niu)] = cardsRet // 牛牛
			} else {
				mapCardTypes[cardType] = cardsRet // 牛1 - 牛9
			}

		}

		cardTypes := storedInt32Slice{}
		for cardType, _ := range mapCardTypes {
			cardTypes = append(cardTypes, cardType)
		}
		cardTypes.Sort()

		cardType := int32(0)
		if len(cardTypes) == 0 {
			// 无牛
			return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}, false
		} else {
			if len(cardTypes) > 1 {
				log.Debugf("牛牛, 顺斗中出现多种普通牌型, cards: %v, cardTypes: %v", cards, cardTypes)
			}
			//cardType 的值是 [1, 10]
			cardType = cardTypes[len(cardTypes)-1]
			return niuniuProto.NiuniuCardType(cardType), mapCardTypes[cardType], true
		}

	}

	return niuniuProto.NiuniuCardType_niuniuCardType_0, []*commonProto.PokerCard{}, false
}

// cards_1 能赢 cards_2 就返回 true, 否则返回 false
func KanCompare(cards_1, cards_2 []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard, niuniuProto.NiuniuCardType, []*commonProto.PokerCard, bool) {
	cardType1, cardComb1 := CalCardType(cards_1)
	kanCardType1, kanCardComb1, _ := Kan(cards_1)
	if kanCardType1 > cardType1 {
		cardType1, cardComb1 = kanCardType1, kanCardComb1
	}

	cardType2, cardComb2 := CalCardType(cards_2)
	kanCardType2, kanCardComb2, _ := Kan(cards_2)
	if kanCardType2 > cardType2 {
		cardType2, cardComb2 = kanCardType2, kanCardComb2
	}

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

// cards_1 能赢 cards_2 就返回 true, 否则返回 false
func ShunCompare(cards_1, cards_2 []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard, niuniuProto.NiuniuCardType, []*commonProto.PokerCard, bool) {
	cardType1, cardComb1 := CalCardType(cards_1)
	cardType2, cardComb2 := CalCardType(cards_2)

	shunCardType1, shunCardComb1, _ := Shun(cards_1)
	if shunCardType1 > cardType1 {
		cardType1, cardComb1 = shunCardType1, shunCardComb1
	}

	shunCardType2, shunCardComb2, _ := Shun(cards_2)
	if shunCardType2 > cardType2 {
		cardType2, cardComb2 = shunCardType2, shunCardComb2
	}

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

// cards_1 能赢 cards_2 就返回 true, 否则返回 false
func KanShunCompare(cards_1, cards_2 []*commonProto.PokerCard) (niuniuProto.NiuniuCardType, []*commonProto.PokerCard, niuniuProto.NiuniuCardType, []*commonProto.PokerCard, bool) {
	cardType1, cardComb1 := CalCardType(cards_1)
	cardType2, cardComb2 := CalCardType(cards_2)
	//1加入坎斗
	kanCardType1, kanCardComb1, _ := Kan(cards_1)
	if kanCardType1 > cardType1 {
		cardType1, cardComb1 = kanCardType1, kanCardComb1
	}
	//1加入顺斗
	shunCardType1, shunCardComb1, _ := Shun(cards_1)
	if shunCardType1 > cardType1 {
		cardType1, cardComb1 = shunCardType1, shunCardComb1
	}
	//2加入顺斗
	shunCardType2, shunCardComb2, _ := Shun(cards_2)
	if shunCardType2 > cardType2 {
		cardType2, cardComb2 = shunCardType2, shunCardComb2
	}
	//2加入坎斗
	kanCardType2, kanCardComb2, _ := Kan(cards_2)
	if kanCardType2 > cardType2 {
		cardType2, cardComb2 = kanCardType2, kanCardComb2
	}

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
