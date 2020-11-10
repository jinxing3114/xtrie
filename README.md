# xtrie

go版本的double array trie算法

增加词等级设置，词等级可有效区分不同等级，不同业务处理模式，词等级支持1-9

压缩数据结构，字符存储更紧凑，占用内存效率更少

扩展多种检索方法
* 词检索
* 前缀检索
* 后缀检索
* 内容检索
* 模糊检索

内容检索和模糊检索的区别在于

内容检索会在内容中查找词库中的词，返回在内容中出现的词

模糊检索会在内容中查找每个词的每个字符，返回内容中出现词的某部分而查找到的词

TO DO list
----------
* 检索性能优化
* 插入性能优化

Reference
---------
[What is Trie](http://en.wikipedia.org/wiki/Trie)   
[An Implementation of Double-Array Trie](http://linux.thai.net/~thep/datrie/datrie.html)

# 快速开始
```sh
go get github.com/jinxing3114/xtrie
```
```go
import "github.com/jinxing3114/xtrie"

var XT = new(xtrie.XTrie)
var storeFile,dictFile = "data/dat.data", "data/darts.txt"
XT.InitHandle(storeFile, dictFile)
```


# 示例
```go
package main

import (
	"fmt"
	"github.com/jinxing3114/xtrie"
)

//创建XTrie
var XT = new(xtrie.XTrie)

func main() {
    var storeFile,dictFile = "data/dat.data", "data/darts.txt"

    XT.InitHandle(storeFile, dictFile)

    //词检索，查找词是否存在于词库中
    index, level, err := XT.Match("a", false)
    fmt.Println(index, level, err)
    
    //文本检索，检索文本中有哪些是词库已存在的词
    content := "文本检索test"
    searchResult := XT.Search(content)
    fmt.Println(searchResult)

    //前缀匹配，根据输入字符，查找满足该前缀的词
    prefixResult,err := XT.Prefix("b", 10)
    fmt.Println(prefixResult, err)

    //后缀匹配，根据输入字符，查找满足该后缀的词
    suffixResult,err := XT.Suffix("c", 10)
    fmt.Println(suffixResult, err)

    //模糊匹配，根据输入的文本任意拆解字符，如果库中存在返回查找到的词
    fuzzyContent := "模糊检索文本测试一下abc"
    fuzzyResult,err := XT.Fuzzy(fuzzyContent, 10)
    fmt.Println(fuzzyResult, err)

    //插入词，复杂度不等，视情况而定
    insertErr := XT.Insert("key", level)
    fmt.Println(insertErr)

    removeErr := XT.Remove("key")
    fmt.Println(removeErr)

}
```

LICENSE
-----------
Apache License 2.0