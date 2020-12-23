package util

import (
	"fmt"
	"os"
)

// WrapErrors 把多个错误合并为一个错误.
func WrapErrors(allErrors ...error) (wrapped error) {
	for _, err := range allErrors {
		if err != nil {
			if wrapped == nil {
				wrapped = err
			} else {
				wrapped = fmt.Errorf("%v | %v", err, wrapped)
			}
		}
	}
	return
}

// UserHomeDir .
func UserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	Panic(err)
	return homeDir
}

// Panic panics if err != nil
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// PathIsNotExist .
func PathIsNotExist(name string) bool {
	_, err := os.Lstat(name)
	if os.IsNotExist(err) {
		return true
	}
	Panic(err)
	return false
}

// PathIsExist .
func PathIsExist(name string) bool {
	return !PathIsNotExist(name)
}

// MustMkdir 确保有一个名为 dirName 的文件夹，
// 如果没有则自动创建，如果已存在则不进行任何操作。
func MustMkdir(dirName string) {
	if PathIsNotExist(dirName) {
		Panic(os.Mkdir(dirName, 0700))
	}
}
