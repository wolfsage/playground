package main

import (
	"fmt"
	"encoding/binary"
	"bytes"
	"os"
	"io/ioutil"
	"crypto/md5"
	"crypto/hmac"
	"encoding/base64"
)

func main() {
	fmt.Println("Opening pkt.pkt (an RNDC packet)")
	f, err := os.Open("pkt.pkt")
	if err != nil {
		fmt.Println(err)
		return
	}

	var pkt []byte
	pkt, _ = ioutil.ReadAll(f)

	var length int32
	var version int32

	buf := bytes.NewBuffer(pkt)

	copy := buf.Bytes()
	buf = bytes.NewBuffer(copy)

	binary.Read(buf, binary.BigEndian, &length)
	binary.Read(buf, binary.BigEndian, &version)
	var x value

	mkey := "abcd"
	ekey, _ := base64.StdEncoding.DecodeString(mkey)

	copy = copy[51:]
	h := hmac.New(md5.New, ekey)
        h.Write(copy)
        sum := h.Sum(nil)
	enc := base64.StdEncoding.EncodeToString(sum)
	enc = enc[:22]

	x = table_fromwire(buf)
	fmt.Println("Decoding packet from disk")
	display(&x, "")

	got := (* (*x.subt)["_auth"].subt)["hmd5"]

	if (got.subv != enc) {
		fmt.Println("Failed to validate packet")
	} else {
		fmt.Println("Packet validated!")
	}

	fmt.Println("Encoding decoded packet")

	// Todo: Properly encode a packet
	zing := table_towire(&x, 0)

	dbytes := zing.Bytes()

	buf2 := bytes.NewBuffer(dbytes)
	var typex int8
	var length2 int32

	binary.Read(buf2, binary.BigEndian, &typex)
	binary.Read(buf2, binary.BigEndian, &length2)
	
	var y value

	fmt.Println("Decoding re-encoded packet")
	y = table_fromwire(buf2)
	display(&y, "")
}

const ISCCC_CCMSGTYPE_STRING     int8 = 0
const ISCCC_CCMSGTYPE_BINARYDATA int8 = 1
const ISCCC_CCMSGTYPE_TABLE      int8 = 2
const ISCCC_CCMSGTYPE_LIST       int8 = 3

type table_head map[string] value
type list_head []value

type value struct {
	xtype int8
	subt *table_head
	suba *list_head
	subv string
}

func value_fromwire (buf *bytes.Buffer) value {
	var typex int8
	var length int32

	binary.Read(buf, binary.BigEndian, &typex)
	binary.Read(buf, binary.BigEndian, &length)

	d := make([]byte, length)
	buf.Read(d)

	nbuf := bytes.NewBuffer(d)

	switch typex {
		case ISCCC_CCMSGTYPE_TABLE:
			return table_fromwire(nbuf)
		case ISCCC_CCMSGTYPE_LIST:
			return list_fromwire(nbuf)
		case ISCCC_CCMSGTYPE_BINARYDATA:
			return binary_fromwire(nbuf)
	}
	var wah value
	return wah
}

func value_towire (val *value) *bytes.Buffer {
	switch val.xtype {
		case ISCCC_CCMSGTYPE_TABLE:
			return table_towire(val,0)
		case ISCCC_CCMSGTYPE_LIST:
			return list_towire(val)
		case ISCCC_CCMSGTYPE_BINARYDATA:
			return binary_towire(val)
	}

	what := new(bytes.Buffer)
	return what
}

func binary_fromwire (buf *bytes.Buffer) value {
	db := make([]byte, buf.Len())
	buf.Read(db)

	var ret value
	ret.xtype = ISCCC_CCMSGTYPE_BINARYDATA
	ret.subv = string(db)
	return ret
}

func binary_towire (val *value) *bytes.Buffer {
	var length int32 = int32(len(val.subv))
	var buf = new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int8(ISCCC_CCMSGTYPE_BINARYDATA))

	if length == 0 {
		binary.Write(buf, binary.BigEndian, int32(4))
		buf.WriteString("null")
	} else {
		binary.Write(buf, binary.BigEndian, length)
		buf.WriteString(val.subv)
	}

	return buf
}		

func list_fromwire (buf *bytes.Buffer) value {
	var data list_head
	for ; buf.Len() > 0; {
		data = append(data, value_fromwire(buf))
	}
	var ret value
	ret.xtype = ISCCC_CCMSGTYPE_LIST
	ret.suba = &data
	return ret
}

func list_towire (val *value) *bytes.Buffer {
	var header = new(bytes.Buffer)
	binary.Write(header, binary.BigEndian, int8(ISCCC_CCMSGTYPE_LIST))
	var buf = new(bytes.Buffer)

	for _, d := range *val.suba {
		buf.Write(value_towire(&d).Bytes())
	}
	
	binary.Write(header, binary.BigEndian, int32(buf.Len()))
	var newbuf = new(bytes.Buffer)
	newbuf.Write(header.Bytes())
	newbuf.Write(buf.Bytes())

	return newbuf
}

func table_fromwire (buf *bytes.Buffer) value {
	data := make(table_head)

	for ; buf.Len() > 0; {
		var length int8
		var key string

		binary.Read(buf, binary.BigEndian, &length)
		d := make([]byte, length)

		buf.Read(d)

		key = string(d)

		data[key] = value_fromwire(buf)
	}

	
	var ret value
	ret.xtype = ISCCC_CCMSGTYPE_TABLE
	ret.subt = &data
	return ret
}

func table_towire (val *value, no_header int) *bytes.Buffer {
	var header = new(bytes.Buffer)
	binary.Write(header, binary.BigEndian, int8(ISCCC_CCMSGTYPE_TABLE))
	var buf = new(bytes.Buffer)

	for k, d := range *val.subt {
		binary.Write(buf, binary.BigEndian, int8(len(k)))
		buf.WriteString(k)
		buf.Write(value_towire(&d).Bytes())
	}

	if (no_header > 0) {
		return buf
	} else {
		var newbuf = new(bytes.Buffer)
		binary.Write(header, binary.BigEndian, int32(buf.Len()))
		newbuf.Write(header.Bytes())
		newbuf.Write(buf.Bytes())
		return newbuf
	}

	return buf
}

func display(v *value, indent string) {
	switch v.xtype {
		case ISCCC_CCMSGTYPE_BINARYDATA:
			fmt.Print("\"", v.subv, "\"\n")
		case ISCCC_CCMSGTYPE_LIST:
			fmt.Print(indent, "(\n");
			cindent := indent
			indent = indent + "  "
			for _, val := range *v.suba {
				display(&val, indent)
			}
			fmt.Print(cindent, ")\n");

		case ISCCC_CCMSGTYPE_TABLE:
			fmt.Print(indent, "{\n");
			cindent := indent
			indent := indent + "  "
			for key, val := range *v.subt {
				fmt.Print(indent, "\"", key, "\":");
				display(&val, indent)
			}
			fmt.Print(cindent, "}\n");
	}
}
