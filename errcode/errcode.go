package errcode

import (
	"google.golang.org/grpc/status"
)

// [http status 3dig] + [biz code 3dig]

const (
	_           = 400000 + iota
	ErrGinBind  // 绑定参数错误
	ErrGinParam // Url参数错误
)

const (
	ErrUnauthorized = 401000 + iota
)

const (
	_                = 403000 + iota
	ErrMutexLock     // 抢锁失败
	ErrMutexUnlock   // 释放锁失败
	ErrBlocked       // 系统己阻断
	ErrKeyDuplicated // 键冲突
	ErrChainFailed   // 链上失败
)

const (
	_                 = 404000 + iota
	ErrRecordNotFound // MySQL记录不存在
	ErrValueNotFound  // Redis值不存在
)

const (
	_                 = 429000 + iota
	ErrRequestBackoff // 请求退避
)

const (
	_                = 500000 + iota
	ErrGinFailed     // Gin内部错误
	ErrRpcTimeout    // Rpc调用超时
	ErrRpcFailed     // Rpc失败
	ErrRpcPanic      // Rpc panic
	ErrSmsSendFailed // Sms发送短信失败
)

func From(err error) (code int, failed bool) {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			code = int(st.Code())
		} else {
			code = 500000
		}
	}

	if code != 0 && (code < 400000 || code >= 500000) {
		failed = true
	}
	return
}
