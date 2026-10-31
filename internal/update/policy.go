package update

import "fmt"

// CheckMinimumNotLowered は最低利用可能バージョンが下がる変更を拒否し、変更前と比べたかどうかを返す
// 変更前の値が版として読めない場合は下げる元の基準が無いので比較しない
// 比較すると、不正な値を直す変更まで拒否して修復できなくなる
func CheckMinimumNotLowered(before, after string) (bool, error) {
	next, err := parseVersion(after)
	if err != nil {
		return false, err
	}

	prev, err := parseVersion(before)
	if err == nil && next.less(prev) {
		return true, fmt.Errorf("%w: %s から %s", ErrMinimumLowered, before, after)
	}

	return err == nil, nil
}
