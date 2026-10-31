package github

import "errors"

var (
	ErrRequestFailed    = errors.New("更新情報の通信に失敗しました")
	ErrUnexpectedStatus = errors.New("更新情報の取得に失敗しました")
	ErrInvalidResponse  = errors.New("更新情報を読めません")
	ErrResponseTooLarge = errors.New("更新情報が大きすぎます")
)
