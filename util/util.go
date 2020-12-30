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

// HasString reports whether item is in the slice.
func HasString(slice []string, item string) bool {
	i := StringIndex(slice, item)
	if i < 0 {
		return false
	}
	return true
}

// StringIndex returns the index of item in the slice.
// returns -1 if not found.
func StringIndex(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// SliceDifference 对比新旧 slice 的差异，并返回需要新增的项目与需要删除的项目。
func SliceDifference(newSlice, oldSlice []string) (toAdd, toDelete []string) {
	// newTags 里有，oldTags 里没有的，需要添加到数据库。
	for _, newItem := range newSlice {
		if !HasString(oldSlice, newItem) {
			toAdd = append(toAdd, newItem)
		}
	}
	// oldTags 里有，newTags 里没有的，需要从数据库中删除。
	for _, oldItem := range oldSlice {
		if !HasString(newSlice, oldItem) {
			toDelete = append(toDelete, oldItem)
		}
	}
	return
}

// DeleteFromSlice .
func DeleteFromSlice(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}
