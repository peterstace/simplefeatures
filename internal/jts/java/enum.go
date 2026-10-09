package java

// Enum is java.lang.Enum, the superclass of generated enum classes.
type Enum struct {
	jself   any
	name    *String
	ordinal int32
}

func (this *Enum) JSelf() any               { return this.jself }
func (this *Enum) JSetSelf(self any)        { this.jself = self }
func (this *Enum) JClass_Enum() *Enum       { return this }
func (this *Enum) Name() *String            { return this.name }
func (this *Enum) Ordinal() int32           { return this.ordinal }
func (this *Enum) ToString() *String        { return this.name }
func (this *Enum) HashCode() int32          { return IdentityHashCode(Ref(this)) }
func (this *Enum) Equals_Object(o any) bool { return Same(Ref(this), o) }

func (this *Enum) CompareTo_Object(o any) int32 {
	return this.ordinal - Enum_fromAny(o).ordinal
}

func Enum_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Enum() *Enum })
	return ok
}

func Enum_fromAny(o any) *Enum {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Enum() *Enum }](o).JClass_Enum()
}

// EnumConst sets the name and ordinal of a newly created enum constant.
func EnumConst[T interface{ JClass_Enum() *Enum }](v T, name string, ordinal int32) T {
	e := v.JClass_Enum()
	e.name = Lit(name)
	e.ordinal = ordinal
	return v
}
