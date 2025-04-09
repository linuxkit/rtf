package local

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Tags are the permitted tags within a test file
type Tags struct {
	Name    string `rt:"NAME"`
	Summary string `rt:"SUMMARY"`
	Author  string `rt:"AUTHOR,allowmultiple"`
	Labels  string `rt:"LABELS"`
	Repeat  int    `rt:"REPEAT"`
	Issue   string `rt:"ISSUE,allowmultiple"`
}

const allowMultiple = "allowmultiple"

func stripOptions(s string) string {
	parts := strings.Split(s, ",")
	return parts[0]
}

func multiplesAllowed(s string) bool {
	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return false
	}
	if parts[1] == allowMultiple {
		return true
	}
	return false
}

// ParseTags reads the provided file and returns all discovered tags or an error
func ParseTags(file string) (*Tags, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	tags := &Tags{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "# ") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) < 2 {
				// Empty
				continue
			}
			tagName := parts[0][2:]
			tagValue := strings.TrimSpace(parts[1])
			tt := reflect.TypeOf(*tags)
			for i := 0; i < tt.NumField(); i++ {
				field := tt.Field(i)
				if rt, ok := field.Tag.Lookup("rt"); ok {
					if stripOptions(rt) == tagName {
						vt := reflect.ValueOf(tags).Elem()
						v := vt.Field(i)
						switch v.Kind() {
						case reflect.Int:
							vi, err := strconv.Atoi(tagValue)
							if err != nil {
								continue
							}
							v.SetInt(int64(vi))
						case reflect.String:
							if multiplesAllowed(rt) {
								if v.String() != "" {
									v.SetString(fmt.Sprintf("%s %s", v.String(), tagValue))
								} else {
									v.SetString(tagValue)
								}
							} else {
								if v.String() != "" {
									return nil, fmt.Errorf("field %s specified multiple times", rt)
								}
								v.SetString(tagValue)
							}
						}
					}
				}
			}
		}
	}
	return tags, nil
}
