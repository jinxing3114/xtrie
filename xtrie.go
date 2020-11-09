// double array 基础和基本方法
// 2020年11月开发。
// 最优的使用场景通过自行维护词典文件来更新词库，该库的优点查询速度快。
// 单个词插入(效率问题，可能会计算调整到整个结构所有的值)
// 删除

package xtrie

import (
	"encoding/gob"
	"errors"
	"log"
	"os"
	"sort"
)

// x trie结构体 double array trie变种结构体
// 数据结构更紧凑
type XTrie struct {
	Fmd5  string  // 词典文件md5
	Size  int     // 切片长度
	Base  []int   // 基础切片，存储字符offset，正值和负值分别代表不同的状态
	Check []int   // 检查字符状态数组，防止查找冲突以及确认多种状态
	Keys  [][]rune// 所有词典转成rune切片
	StoreFile string //dat结构体序列化结果集
	DictFile  string //词典文件路径
	Keymap map[string]int //所有词对应等级
}

//重置基础数据
func (x *XTrie) reset() {
	x.Size = 65535
	x.Fmd5 = ""
	x.Keys = make([][]rune, 0, 1)
	x.Base = make([]int, 0, 65535)
	x.Check  = make([]int, 0, 65535)
	x.Keymap = make(map[string]int)
}

// 重置扩容base和check切片
// 参数 newSize int 新的切片大小
func (x *XTrie) resize(newSize int) int {
	base2  := make([]int, newSize, newSize)
	check2 := make([]int, newSize, newSize)
	if len(x.Base) > 0 {
		copy(base2, x.Base)
		copy(check2, x.Check)
	}
	x.Base  = base2
	x.Check = check2
	x.Size  = newSize
	return newSize
}

// 构造词典，插入词，递归函数，直到找不到下一层深度的词
// 参数 keyPre rune切片 前缀字符切片，查询词等级而设计的
// 参数 children *Node切片 所有子节点
// 参数 index int 上层index值
func (x *XTrie) structure(keyPre []rune, children []*Node, index int) {
	pos := children[0].Code //每次都以字符code起开始查找查找
	//初始查找位置，根据父层索引加子节点最小code算出初始位置
	childLen := len(children)
	offset := 0
outer:
	for {
		pos++
		if pos >= x.Size {
			x.resize(int(float64(offset + children[childLen - 1].Code) * 1.25))
		}
		if x.Base[pos] != 0 {
			continue
		}
		offset = pos - children[0].Code
		if s := offset + children[childLen - 1].Code; s > x.Size { //每次循环计算最大字符code位置是否超出范围
			x.resize(int(float64(s) * 1.25))
		}

		for i := 0; i < childLen; i++ {
			//确保每一个子节点都能落到base和check中
			ind := offset + children[i].Code
			if x.Check[ind] != 0 || x.Base[ind] != 0 {
				continue outer
			}
		}
		break
	}

	if x.Base[index] < 0 { //
		x.Base[index] = offset * 10 + (-x.Base[index])
	} else {
		x.Base[index] = offset
	}

	keyPre = append(keyPre, 1)
	//写入所有的子节点到base中
	//写入所有的子节点到check中
	//必须先把所有节点写入完之后，再去查找添加下一层节点
	for i := 0; i < childLen; i++ {
		ind := offset + children[i].Code
		if children[i].End {
			keyPre[len(keyPre)-1] = rune(children[i].Code)
			x.Base[ind] = -x.Keymap[string(keyPre)]
			x.Check[ind] = -index
		} else {
			x.Base[ind] = 0
			x.Check[ind] = index
		}
	}
	//循环查找下一层节点并且插入dat结构中
	for i := 0; i < childLen; i++ {
		nodes := children[i].fetch(x)
		if len(nodes) > 0 {
			ind := offset + children[i].Code
			keyPre[len(keyPre)-1] = rune(children[i].Code)
			x.structure(keyPre, nodes, ind)
		}
	}
	return
}

// 格式化词库
// 将待格式的词集合，排序之后转为rune字符切片。utf8格式
func (x *XTrie) format() error {
	if len(x.Keymap) == 0 {
		return errors.New("dict is empty")
	}
	allKey := make([]string, 0, len(x.Keymap))
	for k := range x.Keymap {
		allKey = append(allKey, k)
	}
	sort.Strings(allKey)
	x.Keys = make([][]rune, 0, len(allKey))
	for _,key := range allKey {
		x.Keys = append(x.Keys, []rune(key))
	}
	return nil
}

// 编译词库
func (x *XTrie) build () error {
	if len(x.Keymap) == 0 {
		return errors.New("empty Keys")
	}
	err := x.format()
	if err != nil {
		return err
	}
	x.resize(len(x.Keys))
	root := new(Node)
	root.Left = 0
	root.Right = len(x.Keys)
	root.Depth = 0
	children := root.fetch(x)
	rootIndex := 1
	x.structure([]rune{}, children, rootIndex)
	return nil
}

// 使用gob协议
func (x *XTrie) Store(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(x)
	if err != nil {
		return err
	}
	return nil
}

// 从指定路径加载DAT
func (x *XTrie) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Println("dat build file open error", err)
		return err
	}

	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(x)
	if err != nil {
		log.Println("dat build file load error:", err)
		return err
	}

	return nil
}

// 初始化 double array
// 加载store文件，读取词典，编译dat，保存store等
func (x *XTrie) InitHandle(storeFile string, dictFile string) {
	x.StoreFile = storeFile
	x.DictFile  = dictFile
	err := x.Load(x.StoreFile)
	if err != nil { //加载失败
		log.Println("load store", x.StoreFile, "error:", err)
	} else {
		log.Println("load store", x.StoreFile, "success")
	}

	status, err := x.DictRead()
	if status == false {

		if err != nil {
			log.Fatalln("dict file read error:", err)
		} else {
			log.Println("dict file read success")
		}

		err = x.build()
		if err != nil {
			log.Fatalln("build error:", err)
		} else {
			log.Println("build success")
		}

		err = x.Store(x.StoreFile)
		if err != nil {
			log.Fatalln("store error:", err)
		} else {
			log.Println("store save success ")
		}
	} else {
		log.Println("store and dict no difference, do not recompile")
	}
}
