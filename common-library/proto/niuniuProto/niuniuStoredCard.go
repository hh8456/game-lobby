package niuniuProto

import (
	commonProto "servers/common-library/proto/commonProto"
	"sort"
)

type StoredCardSlice []*commonProto.PokerCard

func (s StoredCardSlice) Len() int           { return len(s) }
func (s StoredCardSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StoredCardSlice) Less(i, j int) bool { return s[i].PokerNum < s[j].PokerNum }
func (s StoredCardSlice) Sort()              { sort.Sort(s) }
