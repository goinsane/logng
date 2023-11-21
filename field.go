package logng

// Field is the type of field.
type Field struct {
	Key   string
	Value interface{}
}

// Fields is the slice of fields.
type Fields []Field

// Clone clones the underlying Fields.
func (f Fields) Clone() Fields {
	if f == nil {
		return nil
	}
	f2 := make(Fields, 0, len(f))
	for i := range f {
		f2 = append(f2, f[i])
	}
	return f2
}
