package wherein

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jjeffery/sqlr/private/scanner"
)

var (
	placeholderNumberRE = regexp.MustCompile(`(\?|\$)([0-9]+)?$`)
)

type placeholderInfoT struct {
	leadingSQL        string // SQL before the placeholder
	placeholderText   string
	placeholderPrefix string // prefix that indicates a placeholder ("$", "?")
	origNumber        int    // zero for positional, non-zero for numbered
	argInfo           *argInfoT
}

type argInfoT struct {
	index  int
	offset int
	arg    interface{}
	slice  reflect.Value
	len    int
}

// Expand takes an SQL query and associated arguments and expands out any arguments that
// are a slice of values. Returns the new, expanded SQL query with arguments that have been
// flattened into a slice of scalar argument values.
//
// If args contains only scalar values, then query and args are returned unchanged.
func Expand(query string, args []interface{}) (newQuery string, newArgs []interface{}, err error) {
	if !hasSlice(args) {
		// no changes need to be made
		return query, args, nil
	}

	return flattenQuery(query, args)
}

func flattenQuery(query string, args []interface{}) (newQuery string, newArgs []interface{}, err error) {
	placeholderInfos, trailingSQL, err := newPlaceholderInfos(query)
	if err != nil {
		return "", nil, err
	}

	argInfos := newArgInfos(args)

	numericPlaceholders, err := arePlaceholdersNumeric(placeholderInfos)
	if err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer

	if numericPlaceholders {
		setOffsets(argInfos)
		for _, placeholderInfo := range placeholderInfos {
			buf.WriteString(placeholderInfo.leadingSQL)
			argIndex := placeholderInfo.origNumber - 1
			if argIndex >= len(args) {
				return "", nil, fmt.Errorf("not enough arguments for placeholder %s", placeholderInfo.placeholderText)
			}
			argInfo := argInfos[argIndex]
			start := placeholderInfo.origNumber + argInfo.offset
			count := argInfo.len
			if count == 0 {
				count = 1
			}
			end := start + count
			for n := start; n < end; n++ {
				if n > start {
					buf.WriteRune(',')
				}
				buf.WriteString(placeholderInfo.placeholderPrefix)
				buf.WriteString(strconv.Itoa(n))
			}
		}
	} else {
		if len(argInfos) < len(placeholderInfos) {
			return "", nil, errors.New("not enough arguments for placeholders")
		}
		for i, placeholderInfo := range placeholderInfos {
			buf.WriteString(placeholderInfo.leadingSQL)
			argInfo := argInfos[i]
			if argInfo.len == 0 {
				buf.WriteString(placeholderInfo.placeholderText)
			} else {
				for j := 0; j < argInfo.len; j++ {
					if j > 0 {
						buf.WriteRune(',')
					}
					buf.WriteString(placeholderInfo.placeholderText)
				}
			}
		}
	}

	buf.WriteString(trailingSQL)

	newQuery = buf.String()
	newArgs = flattenArgs(argInfos)
	return newQuery, newArgs, nil
}

func hasSlice(args []interface{}) bool {
	for _, arg := range args {
		switch arg.(type) {
		case string, []byte, int, uint,
			int8, byte,
			int16, uint16,
			int32, uint32,
			int64, uint64,
			float32, float64,
			time.Time,
			driver.Valuer:
			break
		default:
			if rv := reflect.ValueOf(arg); rv.Kind() == reflect.Slice {
				return true
			}
		}
	}
	return false
}

func newPlaceholderInfos(query string) ([]*placeholderInfoT, string, error) {
	var placeholderInfos []*placeholderInfoT
	var buf bytes.Buffer
	scan := scanner.New(strings.NewReader(query))
	for scan.Scan() {
		switch scan.Token() {
		case scanner.PLACEHOLDER:
			submatch := placeholderNumberRE.FindStringSubmatch(scan.Text())
			if submatch == nil {
				return nil, "", fmt.Errorf("unrecognized placeholder %s", scan.Text())
			}
			placeholder := &placeholderInfoT{
				leadingSQL:        buf.String(),
				placeholderPrefix: submatch[1],
				placeholderText:   scan.Text(),
			}
			buf.Reset()
			if len(submatch) > 2 && submatch[2] != "" {
				n, err := strconv.Atoi(submatch[2])
				if err != nil {
					return nil, "", fmt.Errorf("invalid placeholder %s", scan.Text())
				}
				if n < 1 {
					return nil, "", fmt.Errorf("invalid placeholder %s", scan.Text())
				}
				placeholder.origNumber = n
			}
			placeholderInfos = append(placeholderInfos, placeholder)
		case scanner.WS, scanner.COMMENT:
			buf.WriteRune(' ')
		default:
			buf.WriteString(scan.Text())
		}
	}
	//if len(placeholderInfos) == 0 {
	//	return nil, fmt.Errorf("no placeholders in query")
	//}
	return placeholderInfos, buf.String(), nil
}

func arePlaceholdersNumeric(placeholderInfos []*placeholderInfoT) (bool, error) {
	var numericPlaceholders bool
	var positionalPlaceholders bool

	for _, placeholder := range placeholderInfos {
		if placeholder.origNumber > 0 {
			numericPlaceholders = true
		} else {
			positionalPlaceholders = true
		}
	}
	if numericPlaceholders && positionalPlaceholders {
		return false, errors.New("mix of positional and numbered placeholders")
	}

	return numericPlaceholders, nil
}

func newArgInfos(args []interface{}) []*argInfoT {
	argInfos := make([]*argInfoT, 0, len(args))
	for i, arg := range args {
		argInfo := &argInfoT{
			index: i,
			arg:   arg,
		}
		switch arg.(type) {
		case []byte, string,
			int, uint, int8, byte, int16, uint16, int32, uint32, int64, uint64,
			float32, float64, driver.Valuer:
			break
		default:
			if rv := reflect.ValueOf(arg); rv.Kind() == reflect.Slice {
				argInfo.slice = rv
				argInfo.len = rv.Len()
			}
		}
		argInfos = append(argInfos, argInfo)
	}
	return argInfos
}

func flattenArgs(argInfos []*argInfoT) []interface{} {
	var args []interface{}
	for _, argInfo := range argInfos {
		if argInfo.len == 0 {
			// not a slice
			args = append(args, argInfo.arg)
		} else {
			for i := 0; i < argInfo.len; i++ {
				args = append(args, argInfo.slice.Index(i).Interface())
			}
		}
	}
	return args
}

func setOffsets(argInfos []*argInfoT) {
	var offset int
	for _, argInfo := range argInfos {
		argInfo.offset = offset
		if argInfo.len > 0 {
			offset += argInfo.len - 1
		}
	}
}
