// 扩展增加词库管理和检索方法
// 支持如下的检索模式
// Match:完全匹配检索词，判断词是否存在以及获取词相关的数据，复杂度O(n)
// Search:内容检索模式，将内容进行任意拆解组合进行查找词，复杂度O(2n)
// Prefix:前缀检索模式，参数作为前缀，检索相同前缀的词，复杂度O(1)至O(root)
// Suffix:后缀检索模式，参数作为后缀，检索相同后缀的词，复杂度O(1)至O(root)，因检索原理不同，比前缀查找要快
// Fuzzy:模糊查找，将内容进行任意拆解组合，只要词库满足包含某一个字符，检索出相关词
// 支持管理功能：
// Insert:插入词，根据词和词库的匹配程度，复杂度不等，最快O(1)，最慢需要重新构建整个结构
// Delete:删除词，复杂度O(1)
// todo 优化插入词方法，提高检索效率

package xtrie

import (
	"errors"
)

//查询返回值结构体
type MatchResult struct {
	word string
	level int
}

// 判断是否是上下级关系
func (x *XTrie) _upperAndLower(preIndex, index int) error {
	if x.Check[index] != preIndex || -x.Check[index] != preIndex { // 说明上一个字符和当前字符不是上下级关系
		return errors.New("not superior and subordinate")
	}
	return nil
}

// 获取上级索引，本级偏移量，本级词登等级
func (x *XTrie) _getIndexOffset(index int) (preIndex, offset, level int) {
	if x.Base[index] > 0 && x.Check[index] < 0 {
		offset = x.Base[index]/10
		level = x.Base[index]%10
		preIndex = -x.Check[index]
	} else {
		if x.Check[index] > 0 {
			offset = x.Base[index]
			level = 0
			preIndex = x.Check[index]
		} else {
			offset = 0
			level = -x.Base[index]
			preIndex = -x.Check[index]
		}
	}
	return preIndex, offset, level
}

// 查找搜索的词是否在词库-精确查找
// 参数 key string 查找的词
// 返回 最后一个字符的索引，偏移量，等级，error
func (x *XTrie) Match (key string, forceBack bool) (int,int,error) {
	keys, offset, index, level := []rune(key), x.Base[1], 1, 0

	for k,kv := range keys {
		if kv == 0 {
			return index, level, errors.New("code error")
		}
		ind := offset + int(kv)
		if ind >= x.Size { // 超过最大
			return index, level, errors.New("code not set")
		}
		// 说明上一个字符和当前字符不是上下级关系
		if err := x._upperAndLower(index, ind); err != nil{
			return index, level, err
		}
		index = ind
		_, offset, level = x._getIndexOffset(ind)
		if k == len(keys) - 1 {//如果是最后一个字符
			if forceBack { //强制返回模式，一定返回查找到的结果，除非没有结果
				return index, level, nil
			}
			if x.Check[ind] > 0 { //查找到最后一个字符，但是还没到单个词的结尾
				return index, level, errors.New("not found1")
			}
		} else if x.Base[ind] < 0 {
			//如果不是最后一个字符
			//说明没有后续可查的值了，返回查询失败
			return index, level, errors.New("not found2")
		}
	}
	return index, level, nil
}

// 内容匹配模式查找
// 可传入一段文本，逐字查找是否在词库中存在
func (x *XTrie) Search (key string) []MatchResult {
	Keys := []rune(key)
	var start,index,offset,level int
	var result []MatchResult
	for k := range Keys {
		start = -1
		index = 1
		offset = x.Base[index]
		level = 0
		for i:=k;i<len(Keys);i++ {
			//词库没有该字符重置状态继续查找
			ind := offset + int(Keys[i])
			if ind > len(x.Base) { //越界base数组，结束查找
				break
			}
			if err := x._upperAndLower(index, ind); err != nil{
				start = -1
				break
			}
			if start == -1 {
				start = i
			}
			if x.Check[ind] < 0 { //说明该词是结尾标记
				if x.Base[ind] > 0 && x.Check[ind] < 0 {
					level = x.Base[ind] % 10
				} else {
					level = -x.Base[ind]
				}
				result = append(result, MatchResult{string(Keys[start:i + 1]), level})
			}
			if x.Base[ind] < 0 { //如果是结尾状态，没有后续词可查找
				break
			}
			index = ind
			if x.Base[ind] > 0 && x.Check[ind] < 0 {
				offset = x.Base[ind]/10
			} else {
				offset = x.Base[ind]
			}
		}
	}
	return result
}

// 前缀相同字符，添加剩余不同字符
func (x *XTrie) _add(index int, key string,keys []rune, level int) error {
	offset := 0
	preLevel := -x.Base[index]
	for k,v := range keys {
		pos := int(v)
		for {
			pos++
			if pos >= x.Size { //每次循环计算最大字符code位置是否超出范围
				x.resize(int(float64(pos) * 1.25))
			}

			if x.Base[pos] != 0 || x.Check[pos] != 0{
				continue
			}
			offset = pos - int(v)

			if k == len(keys) - 1 {//如果是最后一个字符
				x.Base[index] = offset
				x.Base[pos]   = -level
				x.Check[pos]  = -index
			} else {
				if k == 0 {
					x.Base[index] = offset * 10 + preLevel
				} else {
					x.Base[index] = offset
				}
				x.Base[pos]   = 0
				x.Check[pos]  = index
			}
			index = pos
			break
		}
	}
	_ = x.Store(x.StoreFile)
	x.DictAdd(key, level)
	return nil
}

// 重置树结构，保持最优结构状态
// 移动树结构并添加词
func (x *XTrie) _addMove(keys string, level int) error {

	x.Keymap[keys] = level

	err := x.build()
	if err != nil {
		return errors.New("add key error" + err.Error())
	}

	err = x.Store(x.StoreFile)
	if err != nil {
		return errors.New("add key success, but store DAT is error" + err.Error())
	}

	x.DictAdd(keys, level)

	return nil
}

// 动态添加数据
// 复杂度：可能是O(1)也可能是O(root)
func (x *XTrie) Insert(key string, level int) error {
	//keys := []rune(key)
	//先查找相同前缀的节点
	//获取相同前缀最后的base status，开始添加数据
	//读取已经入库相同前缀的词
	keys, index, offset := []rune(key), 1, x.Base[1]
	//先查找最长的相同前缀index
	for k,kv := range keys {
		if kv == 0 {
			return errors.New("code error")
		}
		ind := offset + int(kv)
		if ind >= x.Size {
			return x._addMove(key, level)
		}
		if x.Check[ind] != index || -x.Check[ind] != index { // 说明上一个字符和当前字符不是上下级关系
			return x._addMove(key, level)
		}
		if k == len(keys) - 1 {//如果是最后一个字符
			if x.Check[ind] > 0 { //查找到最后一个字符，但是还没到单个词的结尾
				x.Check[ind] = -x.Check[ind]
				_ = x.Store(x.StoreFile)
				x.DictAdd(key, level)
			}
			return nil
		} else {
			if x.Base[ind] < 0 { //说明没有后续可查的值了
				return x._add(index, key, keys[k:], level)
			}
			index = ind
			if x.Check[ind] > 0 {
				offset = x.Base[ind]
			} else {
				offset = x.Base[ind]/10
			}
		}
	}
	return nil
}

// 前缀查找，递归方法
func (x *XTrie) _prefix (preStr string, index int, offset int, limit int, result *[]MatchResult) {
	if len(*result) >= limit {//已经查够了不用再查询了
		return
	}
	for i:=2; i<x.Size; i++{
		check := x.Check[i]
		if check != -index && check != index {
			continue
		}
		str := string(rune(i-offset))
		if check < 0 {
			if len(*result) >= limit {
				return
			}
			if x.Base[i] > 0 {
				*result = append(*result, MatchResult{preStr + str, x.Base[i]%10})
			} else {
				*result = append(*result, MatchResult{preStr + str, -x.Base[i]})
			}
		}
		if x.Base[i] > 0 {
			nextOffset := x.Base[i]
			if check < 0 {
				nextOffset = nextOffset/10
			}
			x._prefix(preStr + str, i, nextOffset, limit, result)
		}
	}
}

// 前缀查找
// 匹配搜索词所有相同前缀的词，算法复杂度较高，词不多的时候可以使用
func (x *XTrie) Prefix(pre string, limit int) ([]MatchResult, error) {
	index, level, err := x.Match(pre, true)
	result := make([]MatchResult, 0, limit)
	if err != nil {
		return result, err
	}
	nextOffset := x.Base[index]
	if x.Check[index] < 0 { //说明搜索词是结束词
		result = append(result, MatchResult{pre,level})
		if nextOffset > 0{
			nextOffset = nextOffset/10
		}
	}
	if x.Base[index] < 0 {
		return result, nil
	}
	x._prefix("", index, nextOffset, limit, &result)
	for i:=0;i<len(result);i++ {
		result[i].word = pre + result[i].word
	}
	if len(result) > 10 {
		return result[0:limit], nil
	} else {
		return result, nil
	}
}

// 根据索引查找到的完整前缀字符串
// 调用前缀查找到所有满足相同前缀的字符串
func (x *XTrie) _fuzzyPrefix (index int, off int, limit int, result *[]MatchResult) {
	if len(*result) >= limit {//已经查够了不用再查询了
		return
	}
	//先查找前缀，然后通过前缀匹配查找所有满足条件的词
	key := make([]rune, 0, 5)

	offset, preIndex, nextOffset := 0, index, x.Base[index]

	if nextOffset < 0 { //没有后续的词，结束查找
		*result = append(*result, MatchResult{string(rune(index-off)), -nextOffset})
		return
	} else {
		if x.Check[index] < 0 { //说明搜索词是结束词
			nextOffset = nextOffset/10
		}
	}

	for {
		if preIndex == 1 { //说明已经找到了
			break
		}
		preIndex, offset, _ = x._getIndexOffset(preIndex)
		key = append(key, rune(index-offset))
	}
	x._prefix("", index, nextOffset, limit, result)
}

// 模糊查找
// 命中规则，只要有字符是一样的就会返回，最少一个字符
func (x *XTrie) Fuzzy (key string, limit int) ([]MatchResult, error) {
	keys   := []rune(key)
	result := make([]MatchResult, 0, 10)
	offset := 0
	for i:=2;i<x.Size;i++ {
		_, offset, _ = x._getIndexOffset(i)
		for _,v := range keys {
			//判断是否相同字符
			if int(v) == i-offset {
				x._fuzzyPrefix(i, offset, limit, &result)
				if len(result) >= limit {
					return result, nil
				}
			}
		}
	}
	return result, nil
}

// 后缀匹配词
// 返回查找到的字符串以及词等级
// 算法复杂度，对比前缀搜索要低。根据匹配到的字符依次查找，词越长，查找消耗越大
func (x *XTrie) Suffix (key string, limit int) ([]MatchResult, error) {
	keys        := []rune(key)
	lastRune    := int(keys[len(keys)-1])
	suffixStart := make([]int, 0, 10)
	offset      := 0
	preIndex    := 0
	result      := make([]MatchResult, 0, 10)
	for i:=0;i<x.Size;i++ {
		preIndex = -x.Check[i]
		if preIndex < 0 {
			continue
		}
		if x.Base[preIndex] < 0 {
			offset = -x.Base[preIndex]
		} else {
			offset = x.Base[preIndex]/10
		}
		//判断是否相同结尾字符
		if lastRune != i - offset {
			continue
		}
		level := 0
		if x.Base[i] > 0 {
			level = x.Base[i]/10
		} else {
			level = -x.Base[i]
		}
		if preIndex == 1 {
			result = append(result, MatchResult{string(rune(lastRune)), level})
		} else {
			suffixStart = append(suffixStart, preIndex, level)
		}
	}
	if len(suffixStart) == 0 {
		if len(result) > 0 {
			return result, nil
		} else {
			return result, errors.New("not found")
		}
	}

	for i:=0;i<len(suffixStart);i+=2 {
		index := suffixStart[i]
		z := len(keys) - 2
		c := false
		str := ""
		for {
			if z < 0 { //已找到最后要匹配的字符
				index = 0
				break
			} else {
				if index == 1 { //找到root节点了
					break
				}
				preIndex, offset, _ = x._getIndexOffset(i)
				//判断是否相同结尾字符
				if c == false && int(keys[z]) != index - offset {
					index = 0
					break
				}
				str  += string(rune(index-offset))
				index = preIndex
				if z == 0 { //找到最后一个字符，还没有找到完整的词
					c = true
					continue
				}
				z--
			}

		}
		if index >= 0 { //如果找到词
			result = append(result, MatchResult{string(rune(lastRune))+str, suffixStart[i+1]})
			if len(result) == limit {
				break
			}
		}
	}
	return result, nil
}

// 删除词
// 参数 key string 需要删除的词
func (x *XTrie) Remove(key string) error {

	index, _, err := x.Match(key, false)
	if err != nil {
		return err
	}

	if x.Base[index] > 0 { //该词还有子节点
		if x.Check[index] < 0 { //说明是可结束状态
			x.Check[index] = -x.Check[index]
		}
	} else { //没有子节点，直接清空数据
		x.Base[index]  = 0
		x.Check[index] = 0
	}

	delete(x.Keymap, key)

	err = x.format()
	if err != nil {
		return err
	}

	err = x.Store(x.StoreFile)
	if err != nil {
		return err
	}

	err = x.DictRemove(key)
	return err
}