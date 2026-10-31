package update

import "errors"

var (
	ErrInvalidVersion           = errors.New("バージョン形式が不正です")
	ErrInstallerNotFound        = errors.New("更新用インストーラーが見つかりません")
	ErrMinimumVersionUnreleased = errors.New("最低利用可能バージョンのリリースがありません")
	ErrMinimumLowered           = errors.New("最低利用可能バージョンが下がっています")
)
