package task

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

func GetHash(t Task) string {
	hash := md5.Sum([]byte(PrintTask(t)))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

func PrintTask(t Task) string {
	slices.Sort(t.Tags())
	parts := []string{
		fmt.Sprintf("desc:%s", t.Description()),
		fmt.Sprintf("proj:%s", t.Project()),
		fmt.Sprintf("priority:%s", t.Priority().String()),
		fmt.Sprintf("tags:%s", strings.Join(t.Tags(), ",")),
		fmt.Sprintf("status:%s", t.Status().String()),
	}
	if t.Due() != nil {
		parts = append(parts, fmt.Sprintf("due:%s", t.Due().String()))
	}
	if t.RemotePath() != nil {
		parts = append(parts, fmt.Sprintf("remote:%s", *t.RemotePath()))
	}
	if t.LocalId() != nil {
		parts = append(parts, fmt.Sprintf("local:%s", *t.LocalId()))
	}
	return strings.Join(parts, " ")
}

func Equal(a Task, b Task) bool {
	return GetHash(a) == GetHash(b)
}
