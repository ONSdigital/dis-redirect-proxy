package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/features/steps"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", true, "perform component tests")

type ComponentTest struct {
	RedisFeature *componentTest.RedisFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	ctxBackground := context.Background()
	fmt.Println("starting InitializeScenario")
	f.RedisFeature = componentTest.NewRedisFeature()
	redirectProxyComponent, err := steps.NewProxyComponent(f.RedisFeature)
	if err != nil {
		fmt.Printf("failed to create redirect proxy component - error: %v", err)
		os.Exit(1)
	}
	fmt.Println("In InitializeScenario - calling redirectProxyComponent.InitAPIFeature")
	apiFeature := redirectProxyComponent.InitAPIFeature()

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if f.RedisFeature == nil {
			f.RedisFeature = componentTest.NewRedisFeature()
		}
		apiFeature.Reset()

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		closeErr := f.RedisFeature.Close()
		if closeErr != nil {
			log.Error(ctxBackground, "error occurred while closing the RedisFeature", closeErr)
			os.Exit(1)
		}
		apiFeature.Reset()

		return ctx, nil
	})

	apiFeature.RegisterSteps(ctx)
	f.RedisFeature.RegisterSteps(ctx)
	redirectProxyComponent.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
	ctx.AfterSuite(func() {
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
