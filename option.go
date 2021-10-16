package json

import (
	"github.com/seniorGolang/json/internal/decoder"
	"github.com/seniorGolang/json/internal/encoder"
)

type EncodeOption = encoder.Option
type EncodeOptionFunc func(*EncodeOption)

func UnorderedMap() EncodeOptionFunc {

	return func(opt *EncodeOption) {
		opt.Flag |= encoder.UnorderedMapOption
	}
}

func Debug() EncodeOptionFunc {

	return func(opt *EncodeOption) {
		opt.Flag |= encoder.DebugOption
	}
}

func Colorize(scheme *ColorScheme) EncodeOptionFunc {

	return func(opt *EncodeOption) {
		opt.Flag |= encoder.ColorizeOption
		opt.ColorScheme = scheme
	}
}

type DecodeOption = decoder.Option
type DecodeOptionFunc func(*DecodeOption)

func DecodeFieldPriorityFirstWin() DecodeOptionFunc {
	return func(opt *DecodeOption) {
		opt.Flags |= decoder.FirstWinOption
	}
}
