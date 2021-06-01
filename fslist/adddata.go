package fslist

import (
	"fmt"
	"strings"
	"time"

	"github.com/keyneston/fscache/proto"
	"github.com/rs/zerolog"
)

type AddData struct {
	Name      string
	UpdatedAt *time.Time
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

func (a AddData) MarshalZerologObject(e *zerolog.Event) {
	e.Str("name", a.Name).Bool("isDir", a.IsDir)

	if a.UpdatedAt != nil {
		e.Time("updatedAt", *a.UpdatedAt)
	}
}

func (a AddData) ToProtoFile() *proto.File {
	return &proto.File{
		Dir:  a.IsDir,
		Name: a.Name,
	}
}

func (a AddData) pebbleKey() []byte {
	if len(a.Name) == 0 {
		return []byte{}
	}

	key := []byte(a.Name)
	if a.IsDir && key[len(key)-1] != '/' {
		key = append(key, '/')
	}

	return key
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
