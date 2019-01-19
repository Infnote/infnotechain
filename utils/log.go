package utils

import (
	"fmt"
	"github.com/op/go-logging"
	"os"
	"reflect"
)

var L = logging.MustGetLogger("ifc")

func init() {
	formatter := logging.MustStringFormatter(
		`%{color}[%{time:15:04:05.000}][%{shortfunc}][%{level:.4s}]%{color:reset} %{message}`)
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	formatted := logging.NewBackendFormatter(stdout, formatter)
	leveled := logging.AddModuleLevel(formatted)
	leveled.SetLevel(logging.DEBUG, "")
	logging.SetBackend(leveled)
}

func Dump(i interface{}) string {
	objectValue := reflect.ValueOf(i)
	objectType := objectValue.Type()

	result := fmt.Sprintf("[%v]\n", t.Name())
	for i := 0; i < objectValue.NumField(); i++ {
		field := objectValue.Field(i)
		value := field.Interface()

		output := ""
		switch value.(type) {
		case string:
			output = value.(string)
			if len(t) > 100 {
				output = output[:100]
			}
		case []byte:
			t := v.([]byte)
			if len(t) > 100 {
				t = t[:100]
			}
			output = string(t)
		default:
			output = fmt.Sprintf("%v", Dump(value))
		}

		result += fmt.Sprintf("\t[%v: %v] %v\n", objectType.Field(i).Name, field.Type(), output)
	}

	return result
}
