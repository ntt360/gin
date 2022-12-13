package file

import (
	"encoding/csv"
	"os"
)

// Exists 检测文件是否存在
func Exists(file string) bool {
	info, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// WriteCsv filePath: 文件全路径 绝对路径
// content: 包含行列 需要头部要增加第一列
func WriteCsv(filePath string, content [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入UTF-8 BOM，防止中文乱码
	_, err = file.WriteString("\xEF\xBB\xBF")
	if err != nil {
		return err
	}
	
	w := csv.NewWriter(file)

	for _, v := range content {
		err = w.Write(v)
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}
