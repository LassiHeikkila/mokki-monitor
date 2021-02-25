package influxdb

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type LineProtocol struct {
	Measurement string
	TagSet      map[string]string
	FieldSet    map[string]FieldValue
	Timestamp   time.Time
}

func (l LineProtocol) Length() int {
	totsize := len(l.Measurement)
	for key := range l.TagSet {
		// ",key=value"
		totsize += 1 + len(key) + 1 + len(l.TagSet[key])
	}
	totsize++ // empty space
	for key := range l.FieldSet {
		// "key=value"
		valB, err := l.FieldSet[key].Marshal()
		if err != nil || len(valB) == 0 {
			continue
		}

		totsize += len(key) + 1 + len(valB)
	}
	// commas
	totsize += len(l.FieldSet) - 1

	// empty space
	totsize++
	// unix timestamp
	totsize += len(fmt.Sprintf("%d", l.Timestamp.UnixNano()))
	return totsize
}

func (l LineProtocol) Marshal() ([]byte, error) {
	builder := strings.Builder{}
	builder.Grow(l.Length())

	builder.WriteString(l.Measurement)
	//builder.WriteString("")
	for key := range l.TagSet {
		builder.WriteString(fmt.Sprintf(",%s=%s", key, l.TagSet[key]))
	}
	builder.WriteString(" ")
	n := 0
	for key := range l.FieldSet {
		b, err := l.FieldSet[key].Marshal()
		if err != nil || len(b) == 0 {
			continue
		}
		builder.WriteString(fmt.Sprintf("%s=%s", key, string(b)))
		if n < (len(l.FieldSet) - 1) {
			builder.WriteString(",")
		}
		n++
	}
	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("%d", l.Timestamp.UnixNano()))

	return []byte(builder.String()), nil
}

func (l *LineProtocol) Unmarshal(b []byte) error {
	return errors.New("Unimplemented function")
}
