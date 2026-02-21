package xerrors

import stderrors "errors"

type ErrCode string
type Detail struct {
	Code            ErrCode
	Message         string
	Metadata        map[string]string
	FieldViolations map[string]string
}

type DetailErr struct {
	Err    error
	Detail Detail
}

func (e DetailErr) String() string {
	return e.Err.Error()
}

func (e DetailErr) Unwrap() error {
	return e.Err
}

func (e DetailErr) Error() string {
	return e.Err.Error()
}

// WithDetails оборачивает ошибку с деталью домена.
func WithDetails(err error, detail Detail) error {
	if err == nil {
		err = ErrUnknown
	}
	return DetailErr{Err: err, Detail: detail}
}

// Details извлекает Detail из ошибки (если есть).
func Details(err error) (Detail, bool) {
	var de DetailErr
	if stderrors.As(err, &de) {
		return de.Detail, true
	}
	return Detail{}, false
}
