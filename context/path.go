package context

import (
	"context"
	"path/filepath"

	"github.com/initia-labs/weave/common"
)

func SetInitiaHome(ctx context.Context, initiaHome string) context.Context {
	return context.WithValue(ctx, InitiaHomeKey, initiaHome)
}

func GetInitiaHome(ctx context.Context) string {
	if value, ok := ctx.Value(InitiaHomeKey).(string); ok {
		return value
	}
	panic("cannot cast the InitiaHomeKey value into type string")
}

func GetInitiaConfigDirectory(ctx context.Context) string {
	initiaHome := GetInitiaHome(ctx)
	return filepath.Join(initiaHome, common.InitiaConfigDirectory)
}

func GetInitiaDataDirectory(ctx context.Context) string {
	initiaHome := GetInitiaHome(ctx)
	return filepath.Join(initiaHome, common.InitiaDataDirectory)
}

func SetMinitiaHome(ctx context.Context, minitiaHome string) context.Context {
	return context.WithValue(ctx, MinitiaHomeKey, minitiaHome)
}

func GetMinitiaHome(ctx context.Context) string {
	if value, ok := ctx.Value(MinitiaHomeKey).(string); ok {
		return value
	}
	panic("cannot cast the MinitiaHomeKey value into type string")
}

func GetMinitiaArtifactsConfigJson(ctx context.Context) string {
	minitiaHome := GetMinitiaHome(ctx)
	return filepath.Join(minitiaHome, common.MinitiaArtifactsConfigJson)
}

func GetMinitiaArtifactsJson(ctx context.Context) string {
	minitiaHome := GetMinitiaHome(ctx)
	return filepath.Join(minitiaHome, common.MinitiaArtifactsJson)
}

func SetOPInitHome(ctx context.Context, opInitHome string) context.Context {
	return context.WithValue(ctx, OPInitHomeKey, opInitHome)
}

func GetOPInitHome(ctx context.Context) string {
	if value, ok := ctx.Value(OPInitHomeKey).(string); ok {
		return value
	}
	panic("cannot cast the OPInitHomeKey value into type string")
}
