package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/features/steps"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	RedisFeature          *componentTest.RedisFeature
	ProxiedServiceFeature *steps.ProxiedServiceFeature
	WagtailFeature        *steps.ProxiedServiceFeature
	RedirectProxy         *steps.ProxyComponent
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	// Create shared Redis and mock proxied service
	f.RedisFeature = componentTest.NewRedisFeature()
	f.ProxiedServiceFeature = steps.NewProxiedServiceFeature("Proxied Service")
	f.WagtailFeature = steps.NewProxiedServiceFeature("Wagtail Service")

	// Create the redirect proxy component using those dependencies
	redirectProxyComponent, err := steps.NewProxyComponent(f.RedisFeature, f.ProxiedServiceFeature, f.WagtailFeature)
	if err != nil {
		fmt.Printf("failed to create redirect proxy component - error: %v", err)
		os.Exit(1)
	}
	f.RedirectProxy = redirectProxyComponent

	// Create and attach API feature from the proxy
	apiFeature := f.RedirectProxy.InitAPIFeature()

	// Setup Before hook
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Ensure Redis is ready
		if f.RedisFeature == nil {
			f.RedisFeature = componentTest.NewRedisFeature()
		}

		// Reset API state (routes, recorded requests, etc.)
		apiFeature.Reset()

		return ctx, nil
	})

	// Setup After hook
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Gracefully shut down services
		if f.RedirectProxy != nil {
			_ = f.RedirectProxy.Close()
		}
		if f.RedisFeature != nil {
			_ = f.RedisFeature.Close()
		}

		apiFeature.Reset()
		return ctx, nil
	})

	// Register steps from all components
	apiFeature.RegisterSteps(ctx)
	f.RedisFeature.RegisterSteps(ctx)
	f.ProxiedServiceFeature.RegisterSteps(ctx)
	f.WagtailFeature.RegisterSteps(ctx)
	f.RedirectProxy.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		// No global setup required
	})
	ctx.AfterSuite(func() {
		// No global teardown required
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
			Tags:   "@ReleaseFallback",
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
