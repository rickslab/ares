package logger

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type formatter struct {
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	t := entry.Time.Format("2006/01/02 15:04:05")
	b.WriteString(t)
	b.WriteByte(' ')

	l := strings.ToUpper(entry.Level.String())
	b.WriteString(l)
	b.WriteByte(' ')

	b.WriteString(entry.Message)
	b.WriteByte(' ')

	service := entry.Data["service"]
	method := entry.Data["method"]
	if service != nil {
		b.WriteString(fmt.Sprintf("[%v/%v]", service, method))
	}

	var keys []string
	for k := range entry.Data {
		if k != "service" && k != "method" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	b.WriteByte('(')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}

		v := entry.Data[k]
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(v))
	}
	b.WriteByte(')')
	b.WriteByte('\n')

	return b.Bytes(), nil
}
