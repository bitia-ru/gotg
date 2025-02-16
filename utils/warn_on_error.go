package utils

import "fmt"

func Warn(err error, msg string) {
	if err != nil {
		fmt.Printf("Warning: %s: %v", msg, err)
	}
}
