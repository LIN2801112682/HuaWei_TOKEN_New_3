package matchQuery2

import (
	"dictionary"
	"fmt"
	"github.com/imdario/mergo"
	"index07"
	"reflect"
	"sort"
	_ "strings"
	"time"
)

func MatchSearch(searchStr string, root *dictionary.TrieTreeNode, indexRoot *index07.IndexTreeNode, qmin int, qmax int) []index07.SeriesId {
	//划分查询串为VG
	start1 := time.Now().UnixMicro()
	var vgMap map[int][]string
	vgMap = make(map[int][]string)
	index07.VGCons(root, qmin, qmax, searchStr, vgMap)
	fmt.Println(vgMap)

	//查询每个gram对应倒排个数,并进行排序,把索引项放入sortGramInvertList
	var sortSumInvertList []SortKey
	var sortTokenInvertList []SortTokenInvertList
	for x := range vgMap {
		token := vgMap[x]
		if token != nil {
			invertIndex = nil
			invertIndex2 = nil
			SearchInvertedListFromCurrentNode(token, indexRoot, 0)
			SearchInvertedListFromChildrensOfCurrentNode(indexNode)
			mergo.Merge(&invertIndex, invertIndex2)
			//fmt.Println(len(invertIndex))
			sortSumInvertList = append(sortSumInvertList, NewSortKey(len(invertIndex), token))
			sortTokenInvertList = append(sortTokenInvertList, NewSortTokenInvertList(token, invertIndex))
		}
	}

	sort.SliceStable(sortSumInvertList, func(i, j int) bool {
		if sortSumInvertList[i].sizeOfInvertedList < sortSumInvertList[j].sizeOfInvertedList {
			return true
		}
		return false
	})
	end1 := time.Now().UnixMicro()

	var resArr []index07.SeriesId
	preSeaPosition := 0
	var preInverPositionDis []PosList
	var nowInverPositionDis []PosList
	//fields := strings.Fields(searchStr)
	start2 := time.Now().UnixMicro()
	for m := 0; m < len(sortSumInvertList); m++ {
		tokenArr := sortSumInvertList[m].tokenArr
		var nowSeaPosition int
		if tokenArr != nil {
			for key := range vgMap {
				if reflect.DeepEqual(vgMap[key], tokenArr) {
					nowSeaPosition = key
				}
			}
			invertIndex = nil
			invertIndex2 = nil
			for i := 0; i < len(sortTokenInvertList); i++ {
				if reflect.DeepEqual(sortTokenInvertList[i].tokenArr, tokenArr) {
					invertIndex = sortTokenInvertList[i].indexList
				}
			}
			//fmt.Println(len(invertIndex))
			if invertIndex == nil {
				return nil
			}
			if m == 0 {
				for sid := range invertIndex {
					preInverPositionDis = append(preInverPositionDis, NewPosList(sid, make([]int, len(invertIndex[sid]), len(invertIndex[sid]))))
					nowInverPositionDis = append(nowInverPositionDis, NewPosList(sid, invertIndex[sid]))
					resArr = append(resArr, sid)
				}
			} else {
				for j := 0; j < len(resArr); j++ { //遍历之前合并好的resArr
					findFlag := false
					sid := resArr[j]
					if _, ok := invertIndex[sid]; ok {
						nowInverPositionDis[j] = NewPosList(sid, invertIndex[sid])
						for z1 := 0; z1 < len(preInverPositionDis[j].posArray); z1++ {
							z1Pos := preInverPositionDis[j].posArray[z1]
							for z2 := 0; z2 < len(nowInverPositionDis[j].posArray); z2++ {
								z2Pos := nowInverPositionDis[j].posArray[z2]
								if nowSeaPosition-preSeaPosition == z2Pos-z1Pos {
									findFlag = true
									break
								}
							}
							if findFlag == true {
								break
							}
						}
					}
					if findFlag == false { //没找到并且候选集的sid比resArr大，删除resArr[j]
						resArr = append(resArr[:j], resArr[j+1:]...)
						preInverPositionDis = append(preInverPositionDis[:j], preInverPositionDis[j+1:]...)
						nowInverPositionDis = append(nowInverPositionDis[:j], nowInverPositionDis[j+1:]...)
						j-- //删除后重新指向，防止丢失元素判断
					}
				}
			}
			preSeaPosition = nowSeaPosition
			copy(preInverPositionDis, nowInverPositionDis)
		}
	}
	end2 := time.Now().UnixMicro()
	fmt.Println("精确查询总花费时间（us）：", end2-start1)
	fmt.Println("精确查询划分查询串时间 + 查询索引树 + 排序gram对应倒排表list长度时间（us）：", end1-start1)
	fmt.Println("精确查询合并倒排时间（us）：", end2-start2)
	sort.SliceStable(resArr, func(i, j int) bool {
		if resArr[i].Id() < resArr[j].Id() && resArr[i].Time() < resArr[j].Time() {
			return true
		}
		return false
	})
	return resArr
}

var invertIndex index07.Inverted_index
var indexNode *index07.IndexTreeNode

//查询当前串对应的倒排表（叶子节点）
func SearchInvertedListFromCurrentNode(tokenArr []string, indexRoot *index07.IndexTreeNode, i int) {
	if indexRoot == nil {
		return
	}
	for j := 0; j < len(indexRoot.Children()); j++ {
		if i < len(tokenArr)-1 && tokenArr[i] == indexRoot.Children()[j].Data() {
			SearchInvertedListFromCurrentNode(tokenArr, indexRoot.Children()[j], i+1)
		}
		if i == len(tokenArr)-1 && tokenArr[i] == indexRoot.Children()[j].Data() { //找到那一层的倒排表
			invertIndex = indexRoot.Children()[j].InvertedIndex()
			indexNode = indexRoot.Children()[j]
		}
	}
}

var invertIndex2 index07.Inverted_index

func SearchInvertedListFromChildrensOfCurrentNode(indexNode *index07.IndexTreeNode) {
	if indexNode != nil {
		for l := 0; l < len(indexNode.Children()); l++ {
			if len(indexNode.Children()[l].InvertedIndex()) > 0 {
				mergo.Merge(&invertIndex2, indexNode.Children()[l].InvertedIndex())
			}
			SearchInvertedListFromChildrensOfCurrentNode(indexNode.Children()[l])
		}
	}
}
