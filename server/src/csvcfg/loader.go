package csvcfg

import (
	"encoding/csv"
	"logger"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type tagOptions string

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func LoadConfig(filename string, value interface{}) {
	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		logger.Fatal("value must be ptr %v", v)
	}

	if v.Elem().Kind() != reflect.Map {
		logger.Fatal("value must a map ptr %v", v)
	}

	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		logger.Fatal("Error on LoadConfig for :%v", v.Kind())
	}

	s := v.Elem()

	s.Set(reflect.MakeSlice(v.Elem().Type(), 0, 0))

	infile, err := os.Open(filename)

	r := csv.NewReader(infile)

	r.TrailingComma = true

	out, err := r.ReadAll()

	if err != nil {
		logger.Fatal("Error on loading csv config %s : %s", filename, err.Error())
	}

	if len(out) < 2 {
		logger.Fatal("Error on loading csv : wrong format %s ", filename)
	}

	loadHeader(out)

	et := s.Elem().Type()

	for row := 2; row < len(out); row++ {
		e := reflect.New(et)
		for i := 0; i < e.NumField(); i++ {
			//f := e.Field(i)
			name := et.Field(i).Tag.Get("csv")
			if name == "" {
				name = et.Field(i).Name
			}

			logger.Info("%s %v", name, []byte(name))
		}
	}

	return
}

func getFistUnEmptyValue(records [][]string, row, stop, col int) string {
	if row < 2 {
		logger.Fatal("index wrong col %d", row)
	}

	if stop == 0 {
		stop = 2
	}

	for {
		value := strings.TrimSpace(records[row][col])

		if row == stop {
			return value
		}

		if value != "" {
			return value
		}

		row--
	}

	panic("not be here")
}

func LoadCSVConfig(filename string, value interface{}) {
	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		logger.Fatal("value must be ptr %v", v)
	}

	if v.Elem().Kind() != reflect.Map {
		logger.Fatal("value must a map ptr %v", v)
	}

	mapType := v.Elem().Type()

	switch mapType.Key().Kind() {
	case reflect.String:
	case reflect.Uint32:
	default:
		logger.Fatal("map key must be string or uint32: %v", mapType.Key())
	}

	if mapType.Elem().Kind() != reflect.Ptr {
		logger.Fatal("map value must be ptr: %v", mapType.Elem())
	}

	if mapType.Elem().Elem().Kind() != reflect.Slice {
		logger.Fatal("map value must be ptr to slice: %v", mapType.Elem().Elem())
	}

	sliceType := mapType.Elem().Elem()

	m := v.Elem()

	m.Set(reflect.MakeMap(mapType))

	// 开始正式加载文件
	infile, err := os.Open(filename)

	r := csv.NewReader(infile)

	r.TrailingComma = true

	out, err := r.ReadAll()

	if err != nil {
		logger.Fatal("Error on loading csv config %s : %s", filename, err.Error())
	}

	if len(out) < 2 {
		logger.Fatal("Error on loading csv : wrong format %s , row must >= 2", filename)
	}

	if len(out) == 2 {
		return
	}

	if len(out[0]) < 2 {
		logger.Fatal("Error on loading csv : wrong format %s , col must >= 2", filename)
	}

	header := loadHeader(out)

	keystr := out[2][0]
	var key reflect.Value
	var slice reflect.Value
	var start int

	if strings.TrimSpace(keystr) == "" {
		logger.Fatal("first key must not be space")
	}

	for row := 2; row < len(out); row++ {

		if out[row][0] != "" {
			keystr = strings.TrimSpace(strings.ToLower(out[row][0]))
			start = row

			key = reflect.New(mapType.Key())

			switch key.Elem().Kind() {
			case reflect.String:
				key.Elem().SetString(keystr)
			case reflect.Uint32:
				uv, err := strconv.ParseUint(keystr, 10, 32)
				if err != nil {
					logger.Fatal("can't conv %s to int: %s", keystr, err.Error())
				}

				key.Elem().SetUint(uv)
			default:
				logger.Fatal("error unknow key type %v", key.Elem().Kind())
			}

			slice = reflect.New(sliceType)
			slice.Elem().Set(reflect.MakeSlice(sliceType, 0, 1))
			m.SetMapIndex(key.Elem(), slice)

		}

		if slice.Elem().Len() == slice.Elem().Cap() {
			newcap := slice.Elem().Cap() + slice.Elem().Cap()/2
			if newcap < 4 {
				newcap = 4
			}
			newv := reflect.MakeSlice(sliceType, slice.Elem().Len(), newcap)
			reflect.Copy(newv, slice.Elem())
			slice.Elem().Set(newv)
			//m.SetMapIndex(key.Elem(), slice)
		}

		slice.Elem().SetLen(slice.Elem().Len() + 1)

		et := sliceType.Elem()

		e := reflect.New(et)
		for i := 0; i < et.NumField(); i++ {
			f := e.Elem().Field(i)
			tag := et.Field(i).Tag.Get("csv")
			if tag == "-" {
				continue
			}

			name, opts := parseTag(tag)
			if name == "" {
				name = et.Field(i).Name
			}

			//skip paramter when prefix "_"
			if string(name[0]) == "_" {
				continue
			}

			if idx, exist := header[name]; exist {
				upstart := start
				if opts.Contains("d") {
					upstart = row
				}

				strvalue := getFistUnEmptyValue(out, row, upstart, idx)

				switch f.Kind() {
				case reflect.String:
					f.SetString(strvalue)
				case reflect.Bool:
					if strvalue == "" {
						f.SetBool(false)
					} else {
						b, err := strconv.ParseBool(strvalue)
						if err != nil {
							logger.Fatal("can't conv %s for boolean: %s", strvalue, err.Error())
						}
						f.SetBool(b)
					}

				case reflect.Uint32:
					if strvalue == "" {
						f.SetUint(0)
					} else {
						i, err := strconv.ParseUint(strvalue, 10, 32)
						if err != nil {
							logger.Fatal("can't conv %s for uint32: %s", strvalue, err.Error())
						}
						f.SetUint(i)
					}
				case reflect.Int32:
					if strvalue == "" {
						f.SetInt(0)
					} else {
						i, err := strconv.ParseInt(strvalue, 10, 32)
						if err != nil {
							logger.Fatal("can't conv %s for int32: %s", strvalue, err.Error())
						}
						f.SetInt(i)
					}
				case reflect.Float32:
					if strvalue == "" {
						f.SetFloat(0.0)
					} else {
						rst, err := strconv.ParseFloat(strvalue, 32)
						if err != nil {
							logger.Fatal("can't conv %s for float: ", strvalue, err.Error())
						}
						f.SetFloat(rst)
					}
				default:
					logger.Fatal("unknow type %v", f.Kind())
				}
			} else {
				logger.Fatal("error can't load value for %s", name, filename)
			}
		}

		m.MapIndex(key.Elem()).Elem().Index(slice.Elem().Len() - 1).Set(e.Elem())
	}

	return
}

func loadHeader(records [][]string) (ret map[string]int) {
	ret = make(map[string]int)

	for idx, key := range records[0] {
		ret[key] = idx
	}

	return
}

func GetCfgSize(filename string) int {
	infile, err := os.Open(filename)

	r := csv.NewReader(infile)

	r.TrailingComma = true

	out, err := r.ReadAll()
	if err != nil {
		logger.Error("get config size error", err.Error())
		return 0
	}

	return len(out)
}
