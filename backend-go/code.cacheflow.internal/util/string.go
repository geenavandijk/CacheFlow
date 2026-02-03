package util

import "strings"

// Contains checks if a string is in a slice of strings
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func RemoveStringFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func RemoveStringFromSlicePointers(slice []*string, item string) []*string {
	for i, v := range slice {
		if *v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func ConvertPointerStringArrayToStringArray(ptrArray []*string) []string {
	var strArray []string
	for _, ptr := range ptrArray {
		strArray = append(strArray, *ptr)
	}
	return strArray
}

func ConvertStringArrayToPointerStringArray(strArray []string) []*string {
	var ptrArray []*string
	for _, str := range strArray {
		ptrArray = append(ptrArray, &str)
	}
	return ptrArray
}

func ContainsInt(slice []int32, item int) bool {
	for _, v := range slice {
		if v == int32(item) {
			return true
		}
	}
	return false
}

func Split(s string, sep string) []string {
	return strings.Split(s, sep)
}

// Remove an element from a slice of strings
func Remove(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Check if a string contains another string
func ContainsString(s string, substr string) bool {
	return strings.Contains(s, substr)
}

// Remove an element from a slice of ints
func RemoveInt(slice []int32, item int) []int32 {
	for i, v := range slice {
		if v == int32(item) {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}