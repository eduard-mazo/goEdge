package sparkplug

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Protobuf wire types.
const (
	wireVarint   = 0
	wire64bit    = 1
	wireLenDelim = 2
	wire32bit    = 5
)

// Payload mirrors the Sparkplug B proto2 Payload message (Appendix 1).
type Payload struct {
	Timestamp  uint64
	Metrics    []*Metric
	Seq        uint64
	UUID       string
	Body       []byte
	Properties *PropertySet // field 9: node-level metadata (entity_type, planta_id, fiware_*)
}

// MetaData mirrors the Sparkplug B proto2 MetaData message.
type MetaData struct {
	IsMultiPart bool
	ContentType string
	Size        uint64
	Seq         uint64
	FileName    string
	FileType    string
	MD5         string
	Description string
}

// PropertyValue mirrors the Sparkplug B proto2 PropertyValue message.
type PropertyValue struct {
	Type   uint32
	IsNull bool

	// oneof value
	IntValue    *uint32
	LongValue   *uint64
	FloatValue  *float32
	DoubleValue *float64
	BoolValue   *bool
	StringValue *string
	PropertySet *PropertySet
}

// PropertySet mirrors the Sparkplug B proto2 PropertySet message.
type PropertySet struct {
	Keys   []string
	Values []*PropertyValue
}

// DataSet mirrors the Sparkplug B proto2 DataSet message.
type DataSet struct {
	NumOfColumns uint64
	Columns      []string
	Types        []uint32
	Rows         []*Row
}

// Row mirrors the Sparkplug B proto2 Row message in DataSet.
type Row struct {
	Elements []*DataSetValue
}

// DataSetValue mirrors the Sparkplug B proto2 DataSetValue message.
type DataSetValue struct {
	// oneof value
	IntValue    *uint32
	LongValue   *uint64
	FloatValue  *float32
	DoubleValue *float64
	BoolValue   *bool
	StringValue *string
}

// Template mirrors the Sparkplug B proto2 Template message.
type Template struct {
	Definition   string
	IsDefinition bool
	Metrics      []*Metric
	Parameters   []*Parameter
}

// Parameter mirrors the Sparkplug B proto2 Parameter message in Template.
type Parameter struct {
	Name string
	Type uint32

	// oneof value
	IntValue    *uint32
	LongValue   *uint64
	FloatValue  *float32
	DoubleValue *float64
	BoolValue   *bool
	StringValue *string
}

// Metric mirrors the Sparkplug B proto2 Metric message.
type Metric struct {
	Name      string
	Alias     uint64
	Timestamp uint64
	DataType  uint32

	IsHistorical bool
	IsTransient  bool
	IsNull       bool
	MetaData     *MetaData
	Properties   *PropertySet

	// oneof value
	IntValue      *uint32
	LongValue     *uint64
	FloatValue    *float32
	DoubleValue   *float64
	BoolValue     *bool
	StringValue   *string
	BytesValue    []byte
	DataSetValue  *DataSet
	TemplateValue *Template
}

// Marshal encodes a Payload to Sparkplug B protobuf binary.
func (p *Payload) Marshal() []byte {
	var b []byte
	if p.Timestamp != 0 {
		b = appendTag(b, 1, wireVarint)
		b = appendVarint(b, p.Timestamp)
	}
	for _, m := range p.Metrics {
		enc := m.marshal()
		b = appendTag(b, 2, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	if p.Seq != 0 {
		b = appendTag(b, 3, wireVarint)
		b = appendVarint(b, p.Seq)
	}
	if p.UUID != "" {
		b = appendTag(b, 4, wireLenDelim)
		b = appendVarint(b, uint64(len(p.UUID)))
		b = append(b, []byte(p.UUID)...)
	}
	if len(p.Body) > 0 {
		b = appendTag(b, 5, wireLenDelim)
		b = appendVarint(b, uint64(len(p.Body)))
		b = append(b, p.Body...)
	}
	if p.Properties != nil {
		enc := p.Properties.marshal()
		b = appendTag(b, 9, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

// StringPropertySet builds a PropertySet from an ordered list of key/value pairs.
// Keys with empty values are omitted. Order determines wire order (important for
// readers that zip keys[] with values[] positionally).
func StringPropertySet(keys []string, vals map[string]string) *PropertySet {
	ps := &PropertySet{}
	for _, k := range keys {
		v, ok := vals[k]
		if !ok || v == "" {
			continue
		}
		sv := v
		ps.Keys = append(ps.Keys, k)
		ps.Values = append(ps.Values, &PropertyValue{
			Type:        DataTypeString,
			StringValue: &sv,
		})
	}
	if len(ps.Keys) == 0 {
		return nil
	}
	return ps
}

func (m *Metric) marshal() []byte {
	var b []byte
	if m.Name != "" {
		b = appendTag(b, 1, wireLenDelim)
		b = appendVarint(b, uint64(len(m.Name)))
		b = append(b, []byte(m.Name)...)
	}
	if m.Alias != 0 {
		b = appendTag(b, 2, wireVarint)
		b = appendVarint(b, m.Alias)
	}
	if m.Timestamp != 0 {
		b = appendTag(b, 3, wireVarint)
		b = appendVarint(b, m.Timestamp)
	}
	if m.DataType != 0 {
		b = appendTag(b, 4, wireVarint)
		b = appendVarint(b, uint64(m.DataType))
	}
	if m.IsHistorical {
		b = appendTag(b, 5, wireVarint)
		b = appendVarint(b, 1)
	}
	if m.IsTransient {
		b = appendTag(b, 6, wireVarint)
		b = appendVarint(b, 1)
	}
	if m.IsNull {
		b = appendTag(b, 7, wireVarint)
		b = appendVarint(b, 1)
	}
	if m.MetaData != nil {
		enc := m.MetaData.marshal()
		b = appendTag(b, 8, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	if m.Properties != nil {
		enc := m.Properties.marshal()
		b = appendTag(b, 9, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}

	// oneof value — fields 10-18.
	switch {
	case m.IntValue != nil:
		b = appendTag(b, 10, wireVarint)
		b = appendVarint(b, uint64(*m.IntValue))
	case m.LongValue != nil:
		b = appendTag(b, 11, wireVarint)
		b = appendVarint(b, *m.LongValue)
	case m.FloatValue != nil:
		b = appendTag(b, 12, wire32bit)
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(*m.FloatValue))
		b = append(b, buf[:]...)
	case m.DoubleValue != nil:
		b = appendTag(b, 13, wire64bit)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(*m.DoubleValue))
		b = append(b, buf[:]...)
	case m.BoolValue != nil:
		b = appendTag(b, 14, wireVarint)
		if *m.BoolValue {
			b = appendVarint(b, 1)
		} else {
			b = appendVarint(b, 0)
		}
	case m.StringValue != nil:
		b = appendTag(b, 15, wireLenDelim)
		b = appendVarint(b, uint64(len(*m.StringValue)))
		b = append(b, []byte(*m.StringValue)...)
	case m.BytesValue != nil:
		b = appendTag(b, 16, wireLenDelim)
		b = appendVarint(b, uint64(len(m.BytesValue)))
		b = append(b, m.BytesValue...)
	case m.DataSetValue != nil:
		enc := m.DataSetValue.marshal()
		b = appendTag(b, 17, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	case m.TemplateValue != nil:
		enc := m.TemplateValue.marshal()
		b = appendTag(b, 18, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

func (md *MetaData) marshal() []byte {
	var b []byte
	if md.IsMultiPart {
		b = appendTag(b, 1, wireVarint)
		b = appendVarint(b, 1)
	}
	if md.ContentType != "" {
		b = appendTag(b, 2, wireLenDelim)
		b = appendVarint(b, uint64(len(md.ContentType)))
		b = append(b, []byte(md.ContentType)...)
	}
	if md.Size != 0 {
		b = appendTag(b, 3, wireVarint)
		b = appendVarint(b, md.Size)
	}
	if md.Seq != 0 {
		b = appendTag(b, 4, wireVarint)
		b = appendVarint(b, md.Seq)
	}
	if md.FileName != "" {
		b = appendTag(b, 5, wireLenDelim)
		b = appendVarint(b, uint64(len(md.FileName)))
		b = append(b, []byte(md.FileName)...)
	}
	if md.FileType != "" {
		b = appendTag(b, 6, wireLenDelim)
		b = appendVarint(b, uint64(len(md.FileType)))
		b = append(b, []byte(md.FileType)...)
	}
	if md.MD5 != "" {
		b = appendTag(b, 7, wireLenDelim)
		b = appendVarint(b, uint64(len(md.MD5)))
		b = append(b, []byte(md.MD5)...)
	}
	if md.Description != "" {
		b = appendTag(b, 8, wireLenDelim)
		b = appendVarint(b, uint64(len(md.Description)))
		b = append(b, []byte(md.Description)...)
	}
	return b
}

func (ps *PropertySet) marshal() []byte {
	var b []byte
	for _, k := range ps.Keys {
		b = appendTag(b, 1, wireLenDelim)
		b = appendVarint(b, uint64(len(k)))
		b = append(b, []byte(k)...)
	}
	for _, v := range ps.Values {
		enc := v.marshal()
		b = appendTag(b, 2, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

func (pv *PropertyValue) marshal() []byte {
	var b []byte
	if pv.Type != 0 {
		b = appendTag(b, 1, wireVarint)
		b = appendVarint(b, uint64(pv.Type))
	}
	if pv.IsNull {
		b = appendTag(b, 2, wireVarint)
		b = appendVarint(b, 1)
	}
	switch {
	case pv.IntValue != nil:
		b = appendTag(b, 3, wireVarint)
		b = appendVarint(b, uint64(*pv.IntValue))
	case pv.LongValue != nil:
		b = appendTag(b, 4, wireVarint)
		b = appendVarint(b, *pv.LongValue)
	case pv.FloatValue != nil:
		b = appendTag(b, 5, wire32bit)
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(*pv.FloatValue))
		b = append(b, buf[:]...)
	case pv.DoubleValue != nil:
		b = appendTag(b, 6, wire64bit)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(*pv.DoubleValue))
		b = append(b, buf[:]...)
	case pv.BoolValue != nil:
		b = appendTag(b, 7, wireVarint)
		if *pv.BoolValue {
			b = appendVarint(b, 1)
		} else {
			b = appendVarint(b, 0)
		}
	case pv.StringValue != nil:
		b = appendTag(b, 8, wireLenDelim)
		b = appendVarint(b, uint64(len(*pv.StringValue)))
		b = append(b, []byte(*pv.StringValue)...)
	case pv.PropertySet != nil:
		enc := pv.PropertySet.marshal()
		b = appendTag(b, 9, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

func (ds *DataSet) marshal() []byte {
	var b []byte
	if ds.NumOfColumns != 0 {
		b = appendTag(b, 1, wireVarint)
		b = appendVarint(b, ds.NumOfColumns)
	}
	for _, c := range ds.Columns {
		b = appendTag(b, 2, wireLenDelim)
		b = appendVarint(b, uint64(len(c)))
		b = append(b, []byte(c)...)
	}
	for _, t := range ds.Types {
		b = appendTag(b, 3, wireVarint)
		b = appendVarint(b, uint64(t))
	}
	for _, r := range ds.Rows {
		enc := r.marshal()
		b = appendTag(b, 4, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

func (r *Row) marshal() []byte {
	var b []byte
	for _, e := range r.Elements {
		enc := e.marshal()
		b = appendTag(b, 1, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

func (dv *DataSetValue) marshal() []byte {
	var b []byte
	switch {
	case dv.IntValue != nil:
		b = appendTag(b, 1, wireVarint)
		b = appendVarint(b, uint64(*dv.IntValue))
	case dv.LongValue != nil:
		b = appendTag(b, 2, wireVarint)
		b = appendVarint(b, *dv.LongValue)
	case dv.FloatValue != nil:
		b = appendTag(b, 3, wire32bit)
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(*dv.FloatValue))
		b = append(b, buf[:]...)
	case dv.DoubleValue != nil:
		b = appendTag(b, 4, wire64bit)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(*dv.DoubleValue))
		b = append(b, buf[:]...)
	case dv.BoolValue != nil:
		b = appendTag(b, 5, wireVarint)
		if *dv.BoolValue {
			b = appendVarint(b, 1)
		} else {
			b = appendVarint(b, 0)
		}
	case dv.StringValue != nil:
		b = appendTag(b, 6, wireLenDelim)
		b = appendVarint(b, uint64(len(*dv.StringValue)))
		b = append(b, []byte(*dv.StringValue)...)
	}
	return b
}

func (t *Template) marshal() []byte {
	var b []byte
	if t.Definition != "" {
		b = appendTag(b, 1, wireLenDelim)
		b = appendVarint(b, uint64(len(t.Definition)))
		b = append(b, []byte(t.Definition)...)
	}
	if t.IsDefinition {
		b = appendTag(b, 2, wireVarint)
		b = appendVarint(b, 1)
	}
	for _, m := range t.Metrics {
		enc := m.marshal()
		b = appendTag(b, 3, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	for _, p := range t.Parameters {
		enc := p.marshal()
		b = appendTag(b, 4, wireLenDelim)
		b = appendVarint(b, uint64(len(enc)))
		b = append(b, enc...)
	}
	return b
}

func (p *Parameter) marshal() []byte {
	var b []byte
	if p.Name != "" {
		b = appendTag(b, 1, wireLenDelim)
		b = appendVarint(b, uint64(len(p.Name)))
		b = append(b, []byte(p.Name)...)
	}
	if p.Type != 0 {
		b = appendTag(b, 2, wireVarint)
		b = appendVarint(b, uint64(p.Type))
	}
	switch {
	case p.IntValue != nil:
		b = appendTag(b, 3, wireVarint)
		b = appendVarint(b, uint64(*p.IntValue))
	case p.LongValue != nil:
		b = appendTag(b, 4, wireVarint)
		b = appendVarint(b, *p.LongValue)
	case p.FloatValue != nil:
		b = appendTag(b, 5, wire32bit)
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(*p.FloatValue))
		b = append(b, buf[:]...)
	case p.DoubleValue != nil:
		b = appendTag(b, 6, wire64bit)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(*p.DoubleValue))
		b = append(b, buf[:]...)
	case p.BoolValue != nil:
		b = appendTag(b, 7, wireVarint)
		if *p.BoolValue {
			b = appendVarint(b, 1)
		} else {
			b = appendVarint(b, 0)
		}
	case p.StringValue != nil:
		b = appendTag(b, 8, wireLenDelim)
		b = appendVarint(b, uint64(len(*p.StringValue)))
		b = append(b, []byte(*p.StringValue)...)
	}
	return b
}

// --- Unmarshal implementation ---

func (p *Payload) Unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			p.Timestamp = v
			data = data[n:]
		case 2:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m := &Metric{}
			if err := m.unmarshal(v); err != nil {
				return err
			}
			p.Metrics = append(p.Metrics, m)
			data = data[n:]
		case 3:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			p.Seq = v
			data = data[n:]
		case 4:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			p.UUID = v
			data = data[n:]
		case 5:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			p.Body = v
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (m *Metric) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			m.Name = v
			data = data[n:]
		case 2:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.Alias = v
			data = data[n:]
		case 3:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.Timestamp = v
			data = data[n:]
		case 4:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.DataType = uint32(v)
			data = data[n:]
		case 5:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.IsHistorical = v != 0
			data = data[n:]
		case 6:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.IsTransient = v != 0
			data = data[n:]
		case 7:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.IsNull = v != 0
			data = data[n:]
		case 8:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m.MetaData = &MetaData{}
			if err := m.MetaData.unmarshal(v); err != nil {
				return err
			}
			data = data[n:]
		case 9:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m.Properties = &PropertySet{}
			if err := m.Properties.unmarshal(v); err != nil {
				return err
			}
			data = data[n:]
		case 10:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			uv := uint32(v)
			m.IntValue = &uv
			data = data[n:]
		case 11:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			m.LongValue = &v
			data = data[n:]
		case 12:
			v, n, err := decodeFixed32(data)
			if err != nil {
				return err
			}
			fv := math.Float32frombits(v)
			m.FloatValue = &fv
			data = data[n:]
		case 13:
			v, n, err := decodeFixed64(data)
			if err != nil {
				return err
			}
			dv := math.Float64frombits(v)
			m.DoubleValue = &dv
			data = data[n:]
		case 14:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			bv := v != 0
			m.BoolValue = &bv
			data = data[n:]
		case 15:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			m.StringValue = &v
			data = data[n:]
		case 16:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m.BytesValue = v
			data = data[n:]
		case 17:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m.DataSetValue = &DataSet{}
			if err := m.DataSetValue.unmarshal(v); err != nil {
				return err
			}
			data = data[n:]
		case 18:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m.TemplateValue = &Template{}
			if err := m.TemplateValue.unmarshal(v); err != nil {
				return err
			}
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (md *MetaData) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			md.IsMultiPart = v != 0
			data = data[n:]
		case 2:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			md.ContentType = v
			data = data[n:]
		case 3:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			md.Size = v
			data = data[n:]
		case 4:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			md.Seq = v
			data = data[n:]
		case 5:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			md.FileName = v
			data = data[n:]
		case 6:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			md.FileType = v
			data = data[n:]
		case 7:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			md.MD5 = v
			data = data[n:]
		case 8:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			md.Description = v
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (ps *PropertySet) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			ps.Keys = append(ps.Keys, v)
			data = data[n:]
		case 2:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			pv := &PropertyValue{}
			if err := pv.unmarshal(v); err != nil {
				return err
			}
			ps.Values = append(ps.Values, pv)
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (pv *PropertyValue) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			pv.Type = uint32(v)
			data = data[n:]
		case 2:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			pv.IsNull = v != 0
			data = data[n:]
		case 3:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			uv := uint32(v)
			pv.IntValue = &uv
			data = data[n:]
		case 4:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			pv.LongValue = &v
			data = data[n:]
		case 5:
			v, n, err := decodeFixed32(data)
			if err != nil {
				return err
			}
			fv := math.Float32frombits(v)
			pv.FloatValue = &fv
			data = data[n:]
		case 6:
			v, n, err := decodeFixed64(data)
			if err != nil {
				return err
			}
			dv := math.Float64frombits(v)
			pv.DoubleValue = &dv
			data = data[n:]
		case 7:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			bv := v != 0
			pv.BoolValue = &bv
			data = data[n:]
		case 8:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			pv.StringValue = &v
			data = data[n:]
		case 9:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			pv.PropertySet = &PropertySet{}
			if err := pv.PropertySet.unmarshal(v); err != nil {
				return err
			}
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (ds *DataSet) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			ds.NumOfColumns = v
			data = data[n:]
		case 2:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			ds.Columns = append(ds.Columns, v)
			data = data[n:]
		case 3:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			ds.Types = append(ds.Types, uint32(v))
			data = data[n:]
		case 4:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			r := &Row{}
			if err := r.unmarshal(v); err != nil {
				return err
			}
			ds.Rows = append(ds.Rows, r)
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (r *Row) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			dv := &DataSetValue{}
			if err := dv.unmarshal(v); err != nil {
				return err
			}
			r.Elements = append(r.Elements, dv)
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (dv *DataSetValue) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			uv := uint32(v)
			dv.IntValue = &uv
			data = data[n:]
		case 2:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			dv.LongValue = &v
			data = data[n:]
		case 3:
			v, n, err := decodeFixed32(data)
			if err != nil {
				return err
			}
			fv := math.Float32frombits(v)
			dv.FloatValue = &fv
			data = data[n:]
		case 4:
			v, n, err := decodeFixed64(data)
			if err != nil {
				return err
			}
			fv := math.Float64frombits(v)
			dv.DoubleValue = &fv
			data = data[n:]
		case 5:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			bv := v != 0
			dv.BoolValue = &bv
			data = data[n:]
		case 6:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			dv.StringValue = &v
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (t *Template) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			t.Definition = v
			data = data[n:]
		case 2:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			t.IsDefinition = v != 0
			data = data[n:]
		case 3:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			m := &Metric{}
			if err := m.unmarshal(v); err != nil {
				return err
			}
			t.Metrics = append(t.Metrics, m)
			data = data[n:]
		case 4:
			v, n, err := decodeBytes(data)
			if err != nil {
				return err
			}
			p := &Parameter{}
			if err := p.unmarshal(v); err != nil {
				return err
			}
			t.Parameters = append(t.Parameters, p)
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

func (p *Parameter) unmarshal(data []byte) error {
	for len(data) > 0 {
		fieldNum, wireType, n, err := decodeTag(data)
		if err != nil {
			return err
		}
		data = data[n:]
		switch fieldNum {
		case 1:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			p.Name = v
			data = data[n:]
		case 2:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			p.Type = uint32(v)
			data = data[n:]
		case 3:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			uv := uint32(v)
			p.IntValue = &uv
			data = data[n:]
		case 4:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			p.LongValue = &v
			data = data[n:]
		case 5:
			v, n, err := decodeFixed32(data)
			if err != nil {
				return err
			}
			fv := math.Float32frombits(v)
			p.FloatValue = &fv
			data = data[n:]
		case 6:
			v, n, err := decodeFixed64(data)
			if err != nil {
				return err
			}
			dv := math.Float64frombits(v)
			p.DoubleValue = &dv
			data = data[n:]
		case 7:
			v, n, err := decodeVarint(data)
			if err != nil {
				return err
			}
			bv := v != 0
			p.BoolValue = &bv
			data = data[n:]
		case 8:
			v, n, err := decodeString(data)
			if err != nil {
				return err
			}
			p.StringValue = &v
			data = data[n:]
		default:
			n, err := skipField(wireType, data)
			if err != nil {
				return err
			}
			data = data[n:]
		}
	}
	return nil
}

// --- low-level encoding helpers ---

func appendTag(b []byte, fieldNum, wireType int) []byte {
	return appendVarint(b, uint64(fieldNum<<3|wireType))
}

func appendVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

// --- low-level decoding helpers ---

func decodeTag(data []byte) (fieldNum int, wireType int, n int, err error) {
	v, n, err := decodeVarint(data)
	if err != nil {
		return 0, 0, 0, err
	}
	return int(v >> 3), int(v & 0x7), n, nil
}

func decodeVarint(data []byte) (uint64, int, error) {
	var v uint64
	for i := 0; i < len(data); i++ {
		b := data[i]
		v |= uint64(b&0x7f) << (7 * i)
		if b&0x80 == 0 {
			return v, i + 1, nil
		}
		if i == 9 {
			return 0, 0, errors.New("varint overflow")
		}
	}
	return 0, 0, errors.New("unexpected EOF")
}

func decodeFixed32(data []byte) (uint32, int, error) {
	if len(data) < 4 {
		return 0, 0, errors.New("unexpected EOF")
	}
	return binary.LittleEndian.Uint32(data), 4, nil
}

func decodeFixed64(data []byte) (uint64, int, error) {
	if len(data) < 8 {
		return 0, 0, errors.New("unexpected EOF")
	}
	return binary.LittleEndian.Uint64(data), 8, nil
}

func decodeBytes(data []byte) ([]byte, int, error) {
	v, n, err := decodeVarint(data)
	if err != nil {
		return nil, 0, err
	}
	length := int(v)
	if n+length > len(data) {
		return nil, 0, errors.New("unexpected EOF")
	}
	return data[n : n+length], n + length, nil
}

func decodeString(data []byte) (string, int, error) {
	b, n, err := decodeBytes(data)
	if err != nil {
		return "", 0, err
	}
	return string(b), n, nil
}

func skipField(wireType int, data []byte) (int, error) {
	switch wireType {
	case wireVarint:
		_, n, err := decodeVarint(data)
		return n, err
	case wire64bit:
		if len(data) < 8 {
			return 0, errors.New("unexpected EOF")
		}
		return 8, nil
	case wireLenDelim:
		v, n, err := decodeVarint(data)
		if err != nil {
			return 0, err
		}
		length := int(v)
		if n+length > len(data) {
			return 0, errors.New("unexpected EOF")
		}
		return n + length, nil
	case wire32bit:
		if len(data) < 4 {
			return 0, errors.New("unexpected EOF")
		}
		return 4, nil
	default:
		return 0, fmt.Errorf("unknown wire type %d", wireType)
	}
}

// --- Metric constructors ---

func MetricUInt32(name string, ts uint64, v uint32) *Metric {
	val := v
	return &Metric{Name: name, Timestamp: ts, DataType: DataTypeUInt32, IntValue: &val}
}

func MetricUInt64(name string, ts uint64, v uint64) *Metric {
	val := v
	return &Metric{Name: name, Timestamp: ts, DataType: DataTypeUInt64, LongValue: &val}
}

func MetricDouble(name string, ts uint64, v float64) *Metric {
	val := v
	return &Metric{Name: name, Timestamp: ts, DataType: DataTypeDouble, DoubleValue: &val}
}

func MetricBool(name string, ts uint64, v bool) *Metric {
	val := v
	return &Metric{Name: name, Timestamp: ts, DataType: DataTypeBoolean, BoolValue: &val}
}

func MetricString(name string, ts uint64, v string) *Metric {
	val := v
	return &Metric{Name: name, Timestamp: ts, DataType: DataTypeString, StringValue: &val}
}
