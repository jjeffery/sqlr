package sqlf

import (
	"reflect"
	"strings"
)

// parseTagSettings is adapted from the Gorm
// function of the same name in file model_struct.go.
// Many thanks to the Gorm developers.
// https://github.com/jinzhu/gorm
func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get("gorm")} {
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}
