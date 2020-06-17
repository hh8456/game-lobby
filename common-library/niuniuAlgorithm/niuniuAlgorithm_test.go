package niuniuAlgorithm

import (
	"servers/common-library/log"
	"servers/common-library/proto/commonProto"
	"servers/common-library/proto/niuniuProto"
	"servers/common-library/utility"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	logger := &lumberjack.Logger{
		// 日志输出文件路径
		//Filename: "gate_" + time.Now().Format("2006-01-02.15:04:05") + ".log",
		Filename: "../../log/test.log",
		// 日志文件最大 size, 单位是 MB
		MaxSize: 1000, // megabytes
		// 最大过期日志保留的个数
		MaxBackups: 10,
		// 保留过期文件的最大时间间隔,单位是天
		MaxAge: 28, //days
		// 是否需要压缩滚动日志, 使用的 gzip 压缩
		//Compress: true, // disabled by default
	}
	log.SetOutput(logger) //调用 logrus 的 SetOutput()函数
	log.SetLevel(log.TraceLevel)

}

func calCardId(card *commonProto.PokerCard) int32 {
	return card.PokerColor*100 + card.PokerNum
}

//func TestMakeCardHeap(t *testing.T) {
//for i := 0; i < 100; i++ {
//ids := make([]int32, 0, 52)
//cardHeap := utility.GetPokerHeap()
//for _, card := range cardHeap {
//cardIds := calCardId(card)
//ids = append(ids, cardIds)
//}

//log.Debugf("%v", ids)
//}
//}

//const (
//// 未知类型
//PokerColor_pokerColorNone PokerColor = 0
//// 方块
//PokerColor_pokerColorBox PokerColor = 1
//// 梅花
//PokerColor_pokerColorFlower PokerColor = 2
//// 红桃
//PokerColor_pokerColorHeart PokerColor = 3
//// 黑桃
//PokerColor_pokerColorSpades PokerColor = 4
//)

func TestStoredCard(t *testing.T) {
	pokerCard1 := &commonProto.PokerCard{PokerNum: 4, PokerColor: 1}
	pokerCard2 := &commonProto.PokerCard{PokerNum: 6, PokerColor: 2}
	pokerCard3 := &commonProto.PokerCard{PokerNum: 5, PokerColor: 2}
	pokerCard4 := &commonProto.PokerCard{PokerNum: 11, PokerColor: 4}
	pokerCard5 := &commonProto.PokerCard{PokerNum: 2, PokerColor: 4}

	storedCards := niuniuProto.StoredCardSlice{}
	storedCards = append(storedCards, pokerCard1)
	storedCards = append(storedCards, pokerCard2)
	storedCards = append(storedCards, pokerCard3)
	storedCards = append(storedCards, pokerCard4)
	storedCards = append(storedCards, pokerCard5)

	t.Logf("排序前: %s\n", storedCards)
	storedCards.Sort()
	t.Logf("排序后: %s\n", storedCards)
}

func TestCardTypeShun(t *testing.T) {
	pokerCard := &commonProto.PokerCard{PokerNum: 1, PokerColor: 1}
	pokerCard2 := &commonProto.PokerCard{PokerNum: 2, PokerColor: 2}
	pokerCard3 := &commonProto.PokerCard{PokerNum: 11, PokerColor: 2}
	pokerCard4 := &commonProto.PokerCard{PokerNum: 12, PokerColor: 3}
	pokerCard5 := &commonProto.PokerCard{PokerNum: 13, PokerColor: 4}

	pokerCards := make([]*commonProto.PokerCard, 0, 5)
	pokerCards = append(pokerCards, pokerCard2)
	pokerCards = append(pokerCards, pokerCard)
	pokerCards = append(pokerCards, pokerCard3)
	pokerCards = append(pokerCards, pokerCard4)
	pokerCards = append(pokerCards, pokerCard5)
	// 判断牌型
	cardType, cards, b := Shun(pokerCards)
	if b {
		t.Logf("牌型: %s, 扑克组合: %v",
			niuniuProto.NiuniuCardType(cardType).String(), cards)
	}
}

// 坎斗
func TestCardTypeKan(t *testing.T) {
	pokerCard := &commonProto.PokerCard{PokerNum: 1, PokerColor: 1}
	pokerCard2 := &commonProto.PokerCard{PokerNum: 2, PokerColor: 2}
	pokerCard3 := &commonProto.PokerCard{PokerNum: 11, PokerColor: 2}
	pokerCard4 := &commonProto.PokerCard{PokerNum: 12, PokerColor: 3}
	pokerCard5 := &commonProto.PokerCard{PokerNum: 13, PokerColor: 4}

	pokerCards := make([]*commonProto.PokerCard, 0, 5)
	pokerCards = append(pokerCards, pokerCard2)
	pokerCards = append(pokerCards, pokerCard)
	pokerCards = append(pokerCards, pokerCard3)
	pokerCards = append(pokerCards, pokerCard4)
	pokerCards = append(pokerCards, pokerCard5)
	// 判断牌型
	cardType, cards, b := Kan(pokerCards)
	if b {
		t.Logf("牌型: %s, 扑克组合: %v",
			niuniuProto.NiuniuCardType(cardType).String(), cards)
	}
}

func TestCardTypeOnce(t *testing.T) {
	pokerCard := &commonProto.PokerCard{PokerNum: 4, PokerColor: 1}
	pokerCard2 := &commonProto.PokerCard{PokerNum: 6, PokerColor: 2}
	pokerCard3 := &commonProto.PokerCard{PokerNum: 5, PokerColor: 2}
	pokerCard4 := &commonProto.PokerCard{PokerNum: 11, PokerColor: 4}
	pokerCard5 := &commonProto.PokerCard{PokerNum: 2, PokerColor: 4}

	pokerCards := make([]*commonProto.PokerCard, 0, 5)
	pokerCards = append(pokerCards, pokerCard2)
	pokerCards = append(pokerCards, pokerCard)
	pokerCards = append(pokerCards, pokerCard3)
	pokerCards = append(pokerCards, pokerCard4)
	pokerCards = append(pokerCards, pokerCard5)
	// 判断牌型
	cardType, _ := CalCardType(pokerCards)
	t.Logf("牌型: %s", niuniuProto.NiuniuCardType(cardType).String())
}

func TestCardTypeOnce2(t *testing.T) {
	pokerCard := &commonProto.PokerCard{PokerNum: 7, PokerColor: 3}
	pokerCard2 := &commonProto.PokerCard{PokerNum: 3, PokerColor: 3}
	pokerCard3 := &commonProto.PokerCard{PokerNum: 10, PokerColor: 1}
	pokerCard4 := &commonProto.PokerCard{PokerNum: 11, PokerColor: 3}
	pokerCard5 := &commonProto.PokerCard{PokerNum: 10, PokerColor: 3}

	pokerCards := make([]*commonProto.PokerCard, 0, 5)
	pokerCards = append(pokerCards, pokerCard2)
	pokerCards = append(pokerCards, pokerCard)
	pokerCards = append(pokerCards, pokerCard3)
	pokerCards = append(pokerCards, pokerCard4)
	pokerCards = append(pokerCards, pokerCard5)
	// 判断牌型
	cardType, _ := CalCardType(pokerCards)
	t.Logf("牌型: %s", niuniuProto.NiuniuCardType(cardType).String())
}

func TestCardType(t *testing.T) {
	n := 1000000
	mapType := make(map[int32]int32, 15)
	for i := 0; i < n; i++ {
		// 52 张牌
		cardHeap := utility.GetPokerHeap()
		for j := 0; j < 10; j++ {
			// 判断牌型
			cardType, _ := CalCardType(cardHeap[:5])
			mapType[int32(cardType)]++
			//log.Debugf("%s", cardType.String())
			cardHeap = cardHeap[5:]
		}
	}

	for cardType, num := range mapType {
		log.Debugf("%d 次发牌, %s 出现次数 %d", n*10,
			niuniuProto.NiuniuCardType(cardType).String(), num)
	}
}

func TestNormalCardType(t *testing.T) {
	for i := 0; i < 100000; i++ {
		// 52 张牌
		cardHeap := utility.GetPokerHeap()
		for j := 0; j < 10; j++ {
			// 判断牌型
			cardType, _ := NormalCardType(cardHeap[:5])
			log.Debugf("%s", cardType.String())
			cardHeap = cardHeap[5:]
		}
	}
}
func TestSpecCardType(t *testing.T) {
	for i := 0; i < 1000000; i++ {
		// 52 张牌
		cardHeap := utility.GetPokerHeap()
		for j := 0; j < 10; j++ {
			// 判断牌型
			cardType := SpecialCardType(cardHeap[:5])
			if cardType == niuniuProto.NiuniuCardType_niuniuCardType_seq {
				log.Debugf("%s", cardType.String())
			}

			cardHeap = cardHeap[5:]
		}
	}
}
