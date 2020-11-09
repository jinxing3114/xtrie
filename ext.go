// 扩展double array 方法
// 包括多种查找匹配方法
// 增加和删除等

package xtrie

import (
	"errors"
)

// 查找搜索的词是否在词库-精确查找
// 参数 key string 查找的词
// 返回 最后一个字符的索引，偏移量，等级，error
func (x *XTrie) Match (key string, forceBack bool) (int,int,error) {
	keys, begin, index, level := []rune(key), x.Base[1], 1, 0

	for k,kv := range keys {
		if kv == 0 {
			return index, level, errors.New("code error")
		}
		ind := begin + int(kv)
		abs := x.Check[ind]
		if abs < 0 {
			abs = -abs
		}
		if abs != index { // 说明上一个字符和当前字符不是上下级关系
			return index, level, errors.New("not found key")
		}
		index = ind
		if x.Base[ind] > 0 {
			if x.Check[ind] > 0 {
				begin = x.Base[ind]
				level = 0
			} else {
				begin = x.Base[ind]/10
				level = x.Base[ind]%10
			}
		} else {
			begin = 0
			level = -x.Base[ind]
		}
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
func (x *XTrie) Search (key string) [][3]int {
	Keys := []rune(key)
	var start,index,begin,level int
	var result [][3]int
	for k := range Keys {
		start = -1
		index = 1
		begin = x.Base[index]
		level = 0
		for i:=k;i<len(Keys);i++ {
			//词库没有该字符重置状态继续查找
			ind := begin + int(Keys[i])
			if ind > len(x.Base) { //越界base数组，结束查找
				break
			}
			abs := x.Check[ind]
			if abs < 0 {
				abs = -abs
			}
			if abs != index { // 说明上一个字符和当前字符不是上下级关系
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
				result = append(result, [3]int{start, i, level})
			}
			if x.Base[ind] < 0 { //如果是结尾状态，没有后续词可查找
				break
			}
			index = ind
			if x.Base[ind] > 0 && x.Check[ind] < 0 {
				begin = x.Base[ind]/10
			} else {
				begin = x.Base[ind]
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

/**
	重置树结构，保持最优结构状态
	移动树结构并添加词
	todo 待优化
 */
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
	keys, index, begin := []rune(key), 1, x.Base[1]
	//先查找最长的相同前缀index
	for k,kv := range keys {
		if kv == 0 {
			return errors.New("code error")
		}
		ind := begin + int(kv)
		if ind >= x.Size {
			return x._addMove(key, level)
		}
		abs := x.Check[ind]
		if abs < 0 {
			abs = -abs
		}
		if abs != index { // 说明上一个字符和当前字符不是上下级关系
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
				begin = x.Base[ind]
			} else {
				begin = x.Base[ind]/10
			}
		}
	}
	return nil
}

// 前缀查找，递归方法
func (x *XTrie) _prefix (preStr string, index int, offset int, count *int, limit int) []string {
	result := make([]string, 0, 1)
	if *count >= limit {
		return result
	}
	negative := -index
	for i:=2; i<len(x.Base); i++{
		check := x.Check[i]
		if check != negative && check != index {
			continue
		}
		str := string(rune(i-offset))
		if check < 0 {
			*count++
			result = append(result, preStr + str)
		}
		if x.Base[i] > 0 {
			nextOffset := x.Base[i]
			if check < 0 {
				nextOffset = nextOffset/10
			}
			nextStr := x._prefix(preStr + str, i, nextOffset, count, limit)
			if len(nextStr) > 0 {
				for _,v := range nextStr {
					result = append(result, v)
				}
			}
		}
	}
	return result
}

// 前缀查找
// 匹配搜索词所有相同前缀的词，算法复杂度较高，词不多的时候可以使用
func (x *XTrie) Prefix(pre string, limit int) ([]string, error) {
	index, _, err := x.Match(pre, true)
	result := make([]string, 0, 0)
	if err != nil {
		return result, err
	}
	nextOffset := x.Base[index]
	count := 0
	if x.Check[index] < 0 { //说明搜索词是结束词
		result = append(result, pre)
		if nextOffset > 0{
			nextOffset = nextOffset/10
		}
	}
	if x.Base[index] < 0 {
		return result, nil
	}
	result = append(result, x._prefix("", index, nextOffset, &count, limit)...)
	for i:=0;i<len(result);i++ {
		result[i] = pre + result[i]
	}
	if len(result) > 10 {
		return result[0:limit], nil
	} else {
		return result, nil
	}
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