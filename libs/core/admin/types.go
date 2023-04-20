package admin

type Response struct {
	IsError          bool   `json:"isError"`
	Code             int    `json:"code"`
	TraceId          string `json:"traceId,omitempty"`
	Msg              string `json:"msg"`
	Data             any    `json:"data,omitempty"`
	ValidationErrors any    `json:"validationErrors,omitempty"`
}
