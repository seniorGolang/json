package json

import (
	"github.com/seniorGolang/json/internal/errors"
)

type InvalidUTF8Error = errors.InvalidUTF8Error
type InvalidUnmarshalError = errors.InvalidUnmarshalError
type MarshalerError = errors.MarshalerError
type SyntaxError = errors.SyntaxError
type UnmarshalFieldError = errors.UnmarshalFieldError
type UnmarshalTypeError = errors.UnmarshalTypeError
type UnsupportedTypeError = errors.UnsupportedTypeError
type UnsupportedValueError = errors.UnsupportedValueError
