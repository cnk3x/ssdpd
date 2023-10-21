package flags

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

var (
	CommandLine = flag.CommandLine
)

type FlagSet = flag.FlagSet

type Options struct {
	Prefix      string
	EnvPrefix   string
	Version     string
	Description string
}

func Bind(value any, fOpts *Options) (err error) {
	if value == nil {
		return
	}

	executable, _ := os.Executable()
	CommandLine.Init(strings.TrimSuffix(filepath.Base(executable), ".exe"), flag.ContinueOnError)
	flag.ErrHelp = errors.New("")

	if fOpts == nil {
		fOpts = &Options{}
	}

	rv := reflect.ValueOf(value).Elem()
	rt := rv.Type()

	var flagItems []*fieldValue
	var nl, sl, tl, el int
	var pl = len(fOpts.EnvPrefix)
	var name string

	fieldTag := func(ft reflect.StructField, name string) string {
		return strings.TrimSpace(ft.Tag.Get(name))
	}

	unifiedCase := func(s string) string {
		in := []rune(s)
		isLower := func(idx int) bool {
			return idx >= 0 && idx < len(in) && unicode.IsLower(in[idx])
		}

		out := make([]rune, 0, len(in)+len(in)/2)
		for i, r := range in {
			if unicode.IsUpper(r) {
				r = unicode.ToLower(r)
				if i > 0 && in[i-1] != '-' && (isLower(i-1) || isLower(i+1)) {
					out = append(out, '-')
				}
			}
			out = append(out, r)
		}

		return string(out)
	}

	jEnvKey := func(key string, prefix string) string {
		if key == "" {
			prefix = ""
		}
		return strings.ReplaceAll(prefix+key, "-", "_")
	}

	fieldsSplitRe := regexp.MustCompile(`[\s,;|]+`)

	fieldsSplit := func(s string) (arr []string) {
		var x int
		arr = fieldsSplitRe.Split(s, -1)
		for _, it := range arr {
			if it != "" {
				arr[x] = it
				x++
			}
		}
		arr = arr[:x]
		return
	}

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)

		if !ft.IsExported() {
			continue
		}

		if name = fieldTag(ft, "flag"); name == "" {
			name = unifiedCase(ft.Name)
		}

		if fOpts.Prefix != "" && name != "-" {
			name = fOpts.Prefix + name
		}

		if name = strings.ReplaceAll(name, " ", ""); name == "-" {
			continue
		}

		fit := &fieldValue{Name: name, Value: rv.Field(i), Field: ft}

		if fit.EnvKey = ft.Tag.Get("env"); fit.EnvKey != "-" {
			if fit.EnvKey == "" {
				fit.EnvKey = jEnvKey(strings.ToUpper(fit.Name), fOpts.EnvPrefix)
			} else {
				fit.EnvKey = jEnvKey(fit.EnvKey, fOpts.EnvPrefix)
			}

			if v := os.Getenv(fit.EnvKey); v != "" {
				if err = fit.Set(v); err != nil {
					return
				}
			}
		} else {
			fit.EnvKey = ""
		}

		fit.Short = fieldTag(ft, "short")
		fit.Usage = fieldTag(ft, "usage")
		fit.Alias = fieldsSplit(fieldTag(ft, "alias"))
		flagItems = append(flagItems, fit)

		//apply
		flag.Var(fit, fit.Name, fit.Usage)
		for _, alias := range fit.Alias {
			flag.Var(fit, alias, "alias of "+fit.Name)
		}
		if fit.Short != "" {
			flag.Var(fit, fit.Short, "short of "+fit.Name)
		}

		if l := len(fit.Name) + 2; l > nl {
			nl = l
		}

		for _, n := range fit.Alias {
			if l := len(n) + 2; l > nl {
				nl = l
			}
		}

		if fit.Short != "" {
			if l := len(fit.Short) + 2; l > sl {
				sl = l
			}
		}

		if fit.EnvKey != "" {
			if l := len(fit.EnvKey) + pl + 2; l > el {
				el = l
			}
		}

		if l := len(fit.Type()); l > tl {
			tl = l
		}
	}

	printUsage := func() {
		fmt.Fprint(os.Stderr, CommandLine.Name())
		if fOpts.Version != "" {
			fmt.Fprint(os.Stderr, " - ", fOpts.Version)
		}
		if fOpts.Description != "" {
			fmt.Fprint(os.Stderr, " - ", fOpts.Description)
		}
		fmt.Fprintln(os.Stderr)

		if len(flagItems) == 0 {
			return
		}

		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "命令格式:")
		fmt.Fprintf(os.Stderr, "    %s [...参数选项]\n", CommandLine.Name())
		fmt.Fprintln(os.Stderr)

		fmt.Fprintln(os.Stderr, "参数选项:")
		sort.Slice(flagItems, func(i, j int) bool { return flagItems[i].Name < flagItems[j].Name })
		for _, it := range flagItems {
			fmt.Fprint(os.Stderr, "    ")
			if sl > 0 {
				if it.Short != "" {
					fmt.Fprintf(os.Stderr, "-%*s,", -sl+2, it.Short)
				} else {
					fmt.Fprintf(os.Stderr, "%*s", sl, "")
				}
				fmt.Fprint(os.Stderr, " ")
			}
			fmt.Fprintf(os.Stderr, "--%*s  ", -nl+2, it.Name)
			fmt.Fprintf(os.Stderr, "%*s  ", -tl, it.Type())
			if el > 0 {
				if it.EnvKey != "" {
					fmt.Fprintf(os.Stderr, "%*s", -el, "["+jEnvKey(it.EnvKey, fOpts.EnvPrefix)+"]")
				} else {
					fmt.Fprintf(os.Stderr, "%*s", el, "")
				}
				fmt.Fprint(os.Stderr, " ")
			}
			fmt.Fprint(os.Stderr, it.Usage)
			if vs := it.String(); vs != "" {
				if it.Usage != "" {
					fmt.Fprint(os.Stderr, " ")
				}
				fmt.Fprintf(os.Stderr, `(默认: "%s")`, vs)
			}
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintln(os.Stderr)
	}

	CommandLine.Usage = func() {}

	if err = CommandLine.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr)
		}
		printUsage()
		return
	}

	fromFile := func(fn string, value any) (err error) {
		var data []byte
		if data, err = os.ReadFile(fn); err != nil {
			return
		}
		if ext := strings.ToLower(filepath.Ext(fn)); ext != ".json" {
			var tmp any
			if err = yaml.Unmarshal(data, &tmp); err == nil {
				data, err = json.Marshal(tmp)
			}
			if err != nil {
				return
			}
		}
		err = json.Unmarshal(data, value)
		return
	}

	if hasPath, ok := value.(interface{ ConfigFilePath() string }); ok {
		if configFilePath := hasPath.ConfigFilePath(); configFilePath != "" {
			if err = fromFile(configFilePath, value); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
		}
	}

	return
}

type fieldValue struct {
	Name   string
	Alias  []string
	Short  string
	Usage  string
	EnvKey string
	Value  reflect.Value
	Field  reflect.StructField
}

func (fv *fieldValue) IsBoolFlag() bool { return fv.Field.Type.Kind() == reflect.Bool }

func (fv *fieldValue) String() string {
	vs := reflectGet(fv.Value)
	if len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (fv *fieldValue) Set(s string) error {
	return reflectSet(fv.Value, fv.Field.Type, s)
}

func (fv *fieldValue) Type() string {
	if fv.IsBoolFlag() {
		return ""
	}

	n := strings.ToLower(fv.Field.Type.String())
	for i := len(n) - 1; i >= 0; i-- {
		if n[i] == '.' {
			return n[i+1:]
		}
	}
	return n
}

func reflectGet(fv reflect.Value) (s []string) {
	defer func() { recover() }()

	if !fv.IsValid() {
		return
	}

	switch v := fv.Interface().(type) {
	case time.Duration:
		s = append(s, v.String())
		return
	case time.Time:
		s = append(s, v.Format(time.RFC3339))
		return
	case net.IP:
		if len(v) > 0 {
			s = append(s, v.String())
		}
		return
	case net.IPMask:
		if len(v) > 0 {
			s = append(s, v.String())
		}
		return
	case *net.IPNet:
		if v != nil && len(v.IP) > 0 {
			s = append(s, v.String())
		}
		return
	}

	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s = append(s, strconv.FormatInt(fv.Int(), 10))
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s = append(s, strconv.FormatUint(fv.Uint(), 10))
		return
	case reflect.String:
		s = append(s, fv.String())
		return
	case reflect.Float32, reflect.Float64:
		s = append(s, strconv.FormatFloat(fv.Float(), 'f', 19, 64))
		return
	case reflect.Bool:
		s = append(s, strconv.FormatBool(fv.Bool()))
		return
	case reflect.Slice, reflect.Array:
		for i := 0; i < fv.Len(); i++ {
			s = append(s, reflectGet(fv.Index(i))...)
		}
		return
	}
	return
}

func reflectSet(fv reflect.Value, ft reflect.Type, s string) (err error) {
	defer func() {
		if re := recover(); re != nil {
			if er, ok := re.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("%v", er)
			}
		}
	}()

	if !fv.CanSet() {
		return
	}

	if s == "" && ft.Kind() != reflect.Bool {
		return
	}
	switch fv.Interface().(type) {
	case time.Duration:
		var d time.Duration
		if d, err = time.ParseDuration(s); err != nil {
			return
		}
		fv.SetInt(int64(d))
	case time.Time:
		var t time.Time
		if t, err = time.Parse(time.RFC3339, s); err != nil {
			return
		}
		fv.Set(reflect.ValueOf(t))
	case net.IP, net.IPMask:
		if ip := net.ParseIP(s); ip != nil {
			fv.SetBytes(ip)
		}
	case *net.IPNet:
		if _, _, ni := net.ParseCIDR(s); ni != nil {
			fv.Set(reflect.ValueOf(ni))
		}
	case []byte:
		fv.SetBytes([]byte(s))
	default:
		switch ft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var v int64
			if v, err = strconv.ParseInt(s, 10, 64); err != nil {
				return
			}
			fv.SetInt(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			var v uint64
			if v, err = strconv.ParseUint(s, 10, 64); err != nil {
				return
			}
			fv.SetUint(v)
		case reflect.String:
			fv.SetString(s)
		case reflect.Float32, reflect.Float64:
			var v float64
			if v, err = strconv.ParseFloat(s, 64); err != nil {
				return
			}
			fv.SetFloat(v)
		case reflect.Bool:
			var v = s == ""
			if !v {
				if v, err = strconv.ParseBool(s); err != nil {
					return
				}
			}
			fv.SetBool(v)
		case reflect.Slice, reflect.Array:
			ityp := ft.Elem()
			isPtr := ityp.Kind() == reflect.Pointer
			if isPtr {
				ityp = ityp.Elem()
			}

			iv := reflect.New(ityp)
			if err = reflectSet(iv.Elem(), ityp, s); err != nil {
				return
			}

			if isPtr {
				fv.Set(reflect.Append(fv, iv))
			} else {
				fv.Set(reflect.Append(fv, iv.Elem()))
			}
		}
	}

	return
}
