package opentrace

import "context"

const KeyAction = "action"

func WithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, KeyAction, name)
}
