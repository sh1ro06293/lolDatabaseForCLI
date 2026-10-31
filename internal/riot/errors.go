package riot

import "errors"

var (
	ErrRequestFailed       = errors.New("通信に失敗しました")
	ErrUnexpectedStatus    = errors.New("取得に失敗しました")
	ErrInvalidResponse     = errors.New("応答を読めません")
	ErrUnexpectedFormat    = errors.New("応答の形式が想定と違います")
	ErrEmptyVersions       = errors.New("Data Dragon のバージョン一覧が空です")
	ErrEmptyRoster         = errors.New("Riot のチャンピオン一覧が空です")
	ErrInvalidChampionID   = errors.New("Riot のチャンピオン ID が不正です")
	ErrDuplicateChampionID = errors.New("Riot のチャンピオン ID が重複しています")
)
