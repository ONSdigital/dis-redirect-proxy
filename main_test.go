package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ONSdigital/log.go/v2/log"
	"os"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/features/steps"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", true, "perform component tests")

type ComponentTest struct {
	RedisFeature *componentTest.RedisFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	redirectProxyComponent, err := steps.NewProxyComponent(f.RedisFeature, ctx)
	if err != nil {
		fmt.Printf("failed to create redirect proxy component - error: %v", err)
		os.Exit(1)
	}

	apiFeature := redirectProxyComponent.InitAPIFeature()

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		apiFeature.Reset()
		redirectProxyComponent.Reset()

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if closeErr := redirectProxyComponent.Close(); closeErr != nil {
			panic(closeErr)
		}

		return ctx, nil
	})

	apiFeature.RegisterSteps(ctx)
	//f.RedisFeature.RegisterSteps(ctx)
	redirectProxyComponent.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctxBackground := context.Background()

	ctx.BeforeSuite(func() {
		f.RedisFeature = componentTest.NewRedisFeature()
	})
	ctx.AfterSuite(func() {
		err := f.RedisFeature.Close()
		if err != nil {
			log.Error(ctxBackground, "error occurred while closing the RedisFeature", err)
			os.Exit(1)
		}
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
			Strict: true,
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
