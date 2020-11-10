package xtrie

import (
	"fmt"
	"testing"
)

//创建XTrie
var XT = new(XTrie)

//基本配置信息
var storeFile,dictFile = "data/dat.data", "data/darts.txt"


func TestInit(t *testing.T){

	XT.InitHandle(storeFile, dictFile)
	/**
	例子
	*/
	index, level, err := XT.Match("a", false)
	fmt.Println(index, level, err)

	content := "文本检索kkas"
	searchResult := XT.Search(content)
	fmt.Println(searchResult)

	prefixResult,err := XT.Prefix("b", 10)
	fmt.Println(prefixResult, err)

	suffixResult,err := XT.Suffix("c", 10)
	fmt.Println(suffixResult, err)

	fuzzyContent := "模糊检索文本b测试一下ac"
	fuzzyResult,err := XT.Fuzzy(fuzzyContent, 10)
	fmt.Println(fuzzyResult, err)
	//level, err, _, _ = dat.match("中", false)
	//fmt.Println(level, err)
	//sk := "我要测试一下"
	//contentRune := []rune(sk)
	//result := dat.search(sk)
	//fmt.Println(result)
	//for i:=0;i<len(result);i++{
	//	fmt.Println("str:", string(contentRune[result[i][0]:result[i][1] + 1]), "level", result[i][2])
	//}
}