package model

type ApiResponse struct {
	ErrNo  int    `json:"errno"`
	ErrMsg string `json:"errmsg,omitempty"`
	Data   any    `json:"data,omitempty"`
}

func NewApiResponse(errno int) *ApiResponse {
	return &ApiResponse{
		ErrNo: errno,
	}
}

func (r *ApiResponse) SetErrNo(errno int) *ApiResponse {
	r.ErrNo = errno
	return r
}

func (r *ApiResponse) SetErr(err error) *ApiResponse {
	r.ErrMsg = err.Error()
	return r
}

func (r *ApiResponse) SetErrMsg(errmsg string) *ApiResponse {
	r.ErrMsg = errmsg
	return r
}

func (r *ApiResponse) SetData(data any) *ApiResponse {
	r.Data = data
	return r
}
