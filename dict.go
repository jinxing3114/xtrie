// 词典文件相关操作
// 读取词典
// 删除词
// 添加词

package xtrie

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"strconv"
)

// 读取文件加载词库
func (x *XTrie) DictRead () (bool,error) {
	f, err := os.Open(x.DictFile)
	if err != nil {
		return false, err
	}
	defer f.Close()

	md5h := md5.New()
	_, _ = io.Copy(md5h, f)
	md5Str := hex.EncodeToString(md5h.Sum(nil))
	if x.Fmd5 == md5Str { //如果文件md5一致，不需要重新计算词典
		return true, nil
	} else { //需要重新计算，重置所有状态
		x.reset()
		x.Fmd5 = md5Str
	}

	x.Keymap = make(map[string]int)
	_, _ = f.Seek(0, 0)
	bfRd := bufio.NewReader(f)

	for {
		line, err := bfRd.ReadBytes('\n')
		lineLen := len(line)
		if err != nil && err != io.EOF { //遇到任何错误立即返回，并忽略 EOF 错误信息
			break
		}
		if lineLen >= 3 {//字符太少跳过处理
			if line[lineLen-1] == '\n' {
				lineLen -= 1
			}
			x.Keymap[string(line[2:lineLen])],_ = strconv.Atoi(string(line[0]))
		}
		if err == io.EOF {
			break
		}
	}
	return false, err
}

// 添加词
// todo 写入需要判断最后的字符是不是换行
func (x *XTrie) DictAdd (key string, level int) {
	fd,_:=os.OpenFile(x.DictFile, os.O_RDWR|os.O_CREATE|os.O_APPEND,os.ModePerm)
	_, _ = fd.Write([]byte("\n"+strconv.Itoa(level) + " " + key))
	defer fd.Close()
}

// 移除删除词并将处理后的文件内容写入临时文件中
func _dictRemove(key string, oldDictFile string, tmpDictFile string) error {
	f, err := os.Open(oldDictFile)
	if err != nil {
		return err
	}
	defer f.Close()
	fn, err :=os.OpenFile(tmpDictFile, os.O_RDWR | os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer fn.Close()
	bfRd := bufio.NewReader(f)
	for {
		line, err := bfRd.ReadBytes('\n')
		lineLen := len(line)
		if err != nil && err != io.EOF { //遇到任何错误立即返回，并忽略 EOF 错误信息
			return err
		}
		if lineLen<4 || string(line[2:len(line)-1]) == key {
			if err == io.EOF {
				break
			}
			continue
		}
		_, _ = fn.Write(line)
		if err == io.EOF {
			break
		}
	}
	return nil
}

// 移除词
func (x *XTrie) DictRemove (key string) error {
	err := _dictRemove(key, x.DictFile, x.DictFile+"_tmp")
	if err != nil {
		return err
	}

	err = os.Remove(x.DictFile)
	if err != nil {
		return nil
	}

	err = os.Rename(x.DictFile+"_tmp", x.DictFile)
	if err != nil {
		return err
	}

	return nil
}