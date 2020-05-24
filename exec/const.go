package exec

import (
	"reflect"

	"github.com/qiniu/x/log"
)

// -----------------------------------------------------------------------------

func pushInt32(stk *Context, kind reflect.Kind, v int32) {
	var val interface{}
	switch kind {
	case reflect.Int:
		val = int(v)
	case reflect.Int64:
		val = int64(v)
	case reflect.Int32:
		val = int32(v)
	case reflect.Int16:
		val = int16(v)
	case reflect.Int8:
		val = int8(v)
	default:
		log.Panicln("pushInt failed: invalid kind -", kind)
	}
	stk.Push(val)
}

func pushUint32(stk *Context, kind reflect.Kind, v uint32) {
	var val interface{}
	switch kind {
	case reflect.Uint:
		val = uint(v)
	case reflect.Uint64:
		val = uint64(v)
	case reflect.Uint32:
		val = uint32(v)
	case reflect.Uint8:
		val = uint8(v)
	case reflect.Uint16:
		val = uint16(v)
	case reflect.Uintptr:
		val = uintptr(v)
	default:
		log.Panicln("pushUint failed: invalid kind -", kind)
	}
	stk.Push(val)
}

// -----------------------------------------------------------------------------

var valSpecs = []interface{}{
	false,
	true,
	nil,
}

func execPushValSpec(i Instr, stk *Context) {
	stk.Push(valSpecs[i&bitsOperand])
}

func execPushInt(i Instr, stk *Context) {
	v := int32(i) << bitsOpInt >> bitsOpInt
	kind := reflect.Int + reflect.Kind((i>>bitsOpIntShift)&7)
	pushInt32(stk, kind, v)
}

func execPushUint(i Instr, stk *Context) {
	v := i & bitsOpIntOperand
	kind := reflect.Uint + reflect.Kind((i>>bitsOpIntShift)&7)
	pushUint32(stk, kind, v)
}

func execPushConstR(i Instr, stk *Context) {
	v := stk.code.valConsts[i&bitsOperand]
	stk.Push(v)
}

func execPop(i Instr, stk *Context) {
	n := len(stk.data) - int(i&bitsOperand)
	stk.data = stk.data[:n]
}

// -----------------------------------------------------------------------------

// Push instr
func (p *Builder) pushInstr(val interface{}) (i Instr) {
	if val == nil {
		return iPushNil
	}
	v := reflect.ValueOf(val)
	kind := v.Kind()
	if kind >= reflect.Int && kind <= reflect.Int64 {
		iv := v.Int()
		ivStore := int64(int32(iv) << bitsOpInt >> bitsOpInt)
		if iv == ivStore {
			i = (opPushInt << bitsOpShift) | (uint32(kind-reflect.Int) << bitsOpIntShift) | (uint32(iv) & bitsOpIntOperand)
			return
		}
	} else if kind >= reflect.Uint && kind <= reflect.Uintptr {
		iv := v.Uint()
		if iv == (iv & bitsOpIntOperand) {
			i = (opPushUint << bitsOpShift) | (uint32(kind-reflect.Uint) << bitsOpIntShift) | (uint32(iv) & bitsOpIntOperand)
			return
		}
	} else if kind == reflect.Bool {
		if val.(bool) {
			return iPushTrue
		}
		return iPushFalse
	} else if kind != reflect.String && !(kind >= reflect.Float32 && kind <= reflect.Complex128) {
		log.Panicln("Push failed: unsupported type:", reflect.TypeOf(val), "-", val)
	}
	code := p.code
	i = (opPushConstR << bitsOpShift) | uint32(len(code.valConsts))
	code.valConsts = append(code.valConsts, val)
	return
}

// Push instr
func (p *Builder) Push(val interface{}) *Builder {
	p.code.data = append(p.code.data, p.pushInstr(val))
	return p
}

// Push instr
func (p Reserved) Push(b *Builder, val interface{}) {
	b.code.data[p] = b.pushInstr(val)
}

// Pop instr
func (p *Builder) Pop(n int) *Builder {
	i := (opPop << bitsOpShift) | uint32(n)
	p.code.data = append(p.code.data, i)
	return p
}

// -----------------------------------------------------------------------------
