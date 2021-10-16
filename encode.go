package json

import (
	"context"
	"io"
	"unsafe"

	"github.com/seniorGolang/json/internal/encoder"
	"github.com/seniorGolang/json/internal/encoder/vm"
	"github.com/seniorGolang/json/internal/encoder/vm_color"
	"github.com/seniorGolang/json/internal/encoder/vm_color_indent"
	"github.com/seniorGolang/json/internal/encoder/vm_indent"
)

type Encoder struct {
	w                 io.Writer
	enabledIndent     bool
	enabledHTMLEscape bool
	prefix            string
	indentStr         string
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, enabledHTMLEscape: true}
}

func (e *Encoder) Encode(v interface{}) error {
	return e.EncodeWithOption(v)
}

func (e *Encoder) EncodeWithOption(v interface{}, optFuncs ...EncodeOptionFunc) error {

	ctx := encoder.TakeRuntimeContext()
	ctx.Option.Flag = 0
	err := e.encodeWithOption(ctx, v, optFuncs...)
	encoder.ReleaseRuntimeContext(ctx)
	return err
}

func (e *Encoder) EncodeContext(ctx context.Context, v interface{}, optFuncs ...EncodeOptionFunc) error {

	ctxRt := encoder.TakeRuntimeContext()
	ctxRt.Option.Flag = 0
	ctxRt.Option.Flag |= encoder.ContextOption
	ctxRt.Option.Context = ctx
	err := e.encodeWithOption(ctxRt, v, optFuncs...)
	encoder.ReleaseRuntimeContext(ctxRt)
	return err
}

func (e *Encoder) encodeWithOption(ctx *encoder.RuntimeContext, v interface{}, optFuncs ...EncodeOptionFunc) error {

	if e.enabledHTMLEscape {
		ctx.Option.Flag |= encoder.HTMLEscapeOption
	}
	for _, optFunc := range optFuncs {
		optFunc(ctx.Option)
	}
	var (
		buf []byte
		err error
	)
	if e.enabledIndent {
		buf, err = encodeIndent(ctx, v, e.prefix, e.indentStr)
	} else {
		buf, err = encode(ctx, v)
	}
	if err != nil {
		return err
	}
	if e.enabledIndent {
		buf = buf[:len(buf)-2]
	} else {
		buf = buf[:len(buf)-1]
	}
	buf = append(buf, '\n')
	if _, err := e.w.Write(buf); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) SetEscapeHTML(on bool) {
	e.enabledHTMLEscape = on
}

func (e *Encoder) SetIndent(prefix, indent string) {

	if prefix == "" && indent == "" {
		e.enabledIndent = false
		return
	}
	e.prefix = prefix
	e.indentStr = indent
	e.enabledIndent = true
}

func marshalContext(ctx context.Context, v interface{}, optFuncs ...EncodeOptionFunc) ([]byte, error) {

	ctxRt := encoder.TakeRuntimeContext()
	ctxRt.Option.Flag = 0
	ctxRt.Option.Flag = encoder.HTMLEscapeOption | encoder.ContextOption
	ctxRt.Option.Context = ctx
	for _, optFunc := range optFuncs {
		optFunc(ctxRt.Option)
	}
	buf, err := encode(ctxRt, v)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctxRt)
		return nil, err
	}
	buf = buf[:len(buf)-1]
	copied := make([]byte, len(buf))
	copy(copied, buf)

	encoder.ReleaseRuntimeContext(ctxRt)
	return copied, nil
}

func marshal(v interface{}, optFuncs ...EncodeOptionFunc) ([]byte, error) {

	ctx := encoder.TakeRuntimeContext()
	ctx.Option.Flag = 0
	ctx.Option.Flag |= encoder.HTMLEscapeOption
	for _, optFunc := range optFuncs {
		optFunc(ctx.Option)
	}
	buf, err := encode(ctx, v)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}
	buf = buf[:len(buf)-1]
	copied := make([]byte, len(buf))
	copy(copied, buf)
	encoder.ReleaseRuntimeContext(ctx)
	return copied, nil
}

func marshalNoEscape(v interface{}) ([]byte, error) {

	ctx := encoder.TakeRuntimeContext()
	ctx.Option.Flag = 0
	ctx.Option.Flag |= encoder.HTMLEscapeOption
	buf, err := encodeNoEscape(ctx, v)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}
	buf = buf[:len(buf)-1]
	copied := make([]byte, len(buf))
	copy(copied, buf)

	encoder.ReleaseRuntimeContext(ctx)
	return copied, nil
}

func marshalIndent(v interface{}, prefix, indent string, optFuncs ...EncodeOptionFunc) ([]byte, error) {

	ctx := encoder.TakeRuntimeContext()
	ctx.Option.Flag = 0
	ctx.Option.Flag |= encoder.HTMLEscapeOption | encoder.IndentOption
	for _, optFunc := range optFuncs {
		optFunc(ctx.Option)
	}
	buf, err := encodeIndent(ctx, v, prefix, indent)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}
	buf = buf[:len(buf)-2]
	copied := make([]byte, len(buf))
	copy(copied, buf)
	encoder.ReleaseRuntimeContext(ctx)
	return copied, nil
}

func encode(ctx *encoder.RuntimeContext, v interface{}) ([]byte, error) {

	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendComma(ctx, b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ
	typePtr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encoder.CompileToGetCodeSet(typePtr)
	if err != nil {
		return nil, err
	}
	p := uintptr(header.ptr)
	ctx.Init(p, codeSet.CodeLength)
	ctx.KeepRefs = append(ctx.KeepRefs, header.ptr)
	buf, err := encodeRunCode(ctx, b, codeSet)
	if err != nil {
		return nil, err
	}
	ctx.Buf = buf
	return buf, nil
}

func encodeNoEscape(ctx *encoder.RuntimeContext, v interface{}) ([]byte, error) {

	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendComma(ctx, b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ
	typePtr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encoder.CompileToGetCodeSet(typePtr)
	if err != nil {
		return nil, err
	}
	p := uintptr(header.ptr)
	ctx.Init(p, codeSet.CodeLength)
	buf, err := encodeRunCode(ctx, b, codeSet)
	if err != nil {
		return nil, err
	}
	ctx.Buf = buf
	return buf, nil
}

func encodeIndent(ctx *encoder.RuntimeContext, v interface{}, prefix, indent string) ([]byte, error) {

	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(ctx, b)
		b = encoder.AppendCommaIndent(ctx, b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ
	typePtr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encoder.CompileToGetCodeSet(typePtr)
	if err != nil {
		return nil, err
	}
	p := uintptr(header.ptr)
	ctx.Init(p, codeSet.CodeLength)
	buf, err := encodeRunIndentCode(ctx, b, codeSet, prefix, indent)
	ctx.KeepRefs = append(ctx.KeepRefs, header.ptr)
	if err != nil {
		return nil, err
	}
	ctx.Buf = buf
	return buf, nil
}

func encodeRunCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet) ([]byte, error) {

	if (ctx.Option.Flag & encoder.DebugOption) != 0 {
		if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
			return vm_color.DebugRun(ctx, b, codeSet)
		}
		return vm.DebugRun(ctx, b, codeSet)
	}
	if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
		return vm_color.Run(ctx, b, codeSet)
	}
	return vm.Run(ctx, b, codeSet)
}

func encodeRunIndentCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, prefix, indent string) ([]byte, error) {

	ctx.Prefix = []byte(prefix)
	ctx.IndentStr = []byte(indent)
	if (ctx.Option.Flag & encoder.DebugOption) != 0 {
		if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
			return vm_color_indent.DebugRun(ctx, b, codeSet)
		}
		return vm_indent.DebugRun(ctx, b, codeSet)
	}
	if (ctx.Option.Flag & encoder.ColorizeOption) != 0 {
		return vm_color_indent.Run(ctx, b, codeSet)
	}
	return vm_indent.Run(ctx, b, codeSet)
}
