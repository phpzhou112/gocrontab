package tools

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var cfgCache = make(map[string]string)

//解析配置文件
func Get(conf string, path string, flag bool) gjson.Result {
	var filename string
	if flag {
		filename = "manager/config/" + conf + ".conf"
	}else{
		filename = "worker/config/" + conf + ".conf"
	}

	content := CfgRead(filename)
	if !gjson.Valid(content) {
		return gjson.Result{}
	}

	return gjson.Get(content, path)
}

//验证配置文件
func Valid(filename string) {
	filenameWithSuffix := path.Base(filename)  //获取文件名带后缀
	fileSuffix := path.Ext(filenameWithSuffix) //获取文件后缀
	if fileSuffix == ".conf" {
		content := CfgRead(filename)
		if !gjson.Valid(content) {
			//配置文件格式不正确，请检查配置文件
			panic("配置文件格式不正确，请检查配置文件:" + filename)
		}
	}

	return
}

func CfgRead(filename string) string {
	content := cacheGet(filename)
	if content != "" {
		return content
	} else {
		ret, _ := cacheIn(filename)

		if ret == false {
			panic("config file not found:" + filename)
		}

		data, e := ioutil.ReadFile(filename)
		if e != nil {
			panic("read config file {" + filename + "} error:" + e.Error())
		}

		content = string(data)
		cacheSet(filename, content)
	}

	return content
}

func cacheGet(key string) string {
	value, exists := cfgCache[key]
	if exists {
		return value
	}

	return ""
}

func cacheSet(key string, val string) {
	_, exists := cfgCache[key]
	if exists {
		cfgCache[key] = val
		return
	}

	cfgCache[key] = val
}

func cacheIn(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil { //忽略错误
			return err
		}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}
