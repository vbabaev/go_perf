package report

import (
    "bytes"
    "encoding/binary"
    "time"
)

const (
    TYPE_UNKNOWN = iota
    TYPE_CPU = iota
    TYPE_MEM = iota
    TYPE_PROC = iota
)

type (
    Report struct {
        _time uint32
        _type uint32
        _text string
        _packed *[]byte
    }
)

func Create(type_ uint32, text string) Report {
    time := uint32(time.Now().Unix())
    return Report{time, type_, text, nil}
}

func Load(time uint32, type_ uint32, binary_message []byte) Report{
    message := string(binary_message)
    return Report{time, type_, message, nil}
}

func (r *Report) Pack () []byte {
    var packed []byte
    if r._packed == nil {
        header_size := GetHeaderSize()
        size := header_size + uint32(len(r._text))
        packed = make([]byte, size, size)
        copy(packed[0:header_size], r.header())
        copy(packed[header_size:], r._text)
        r._packed = &packed
    }
    return packed
}

func (r *Report) header() []byte {
    size := GetHeaderSize()
    header := make([]byte, size, size)
    binary.LittleEndian.PutUint32(header[0:4], uint32(len(r._text)))
    binary.LittleEndian.PutUint32(header[4:8], r._type)
    binary.LittleEndian.PutUint32(header[8:12], r._time)
    return header
}

func (r *Report) Time() uint32 {
    return r._time
}

func (r *Report) Text() string {
    return r._text
}

func GetHeaderSize() uint32 {
    // 4 - for time uint32
    // 4 - for type uint32
    // 4 - for length uint32
    return 4 + 4 + 4
}

func ParseHeader(header []byte) (time, length, type_ uint32) {
    length = read_uint32(header[0:4])
    type_ = read_uint32(header[4:8])
    time = read_uint32(header[8:12])
    return 
}

func read_uint32(data []byte) (ret uint32) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}
