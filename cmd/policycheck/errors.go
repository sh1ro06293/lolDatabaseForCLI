package main

import "errors"

var ErrUsage = errors.New("使い方: policycheck <変更前の update-policy.json> <変更後の update-policy.json>")
