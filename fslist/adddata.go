package fslist

import (
	"fmt"
	"strings"
	"time"

	"github.com/keyneston/fscache/proto"
)

type AddData struct {
	Name      string
	UpdatedAt time.Time
	IsDir     bool
}

func AddDataFromProtoFile(f *proto.File) AddData {
	return AddData{
		Name:  f.Name,
		IsDir: f.Dir,
	}
}

func (a AddData) String() string {
	dirStr := ""
	if a.IsDir {
		dirStr = "/"
	}

	return fmt.Sprintf("AddData{%s%s}", a.Name, dirStr)
}

func (a AddData) ToProtoFile() *proto.File {
	return &proto.File{
		Dir:  a.IsDir,
		Name: a.Name,
	}
}

type ByPath []AddData

func (s ByPath) Len() int {
	return len(s)
}
func (s ByPath) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByPath) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}
