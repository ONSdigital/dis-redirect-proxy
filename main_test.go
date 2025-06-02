package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/features/steps"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	redirectProxyComponent, err := steps.NewRedirectProxyComponent()
	if err != nil {
		fmt.Printf("failed to create redirect proxy component - error: %v", err)
		os.Exit(1)
	}

	apiFeature := redirectProxyComponent.InitAPIFeature()

	url := fmt.Sprintf("http://%s", redirectProxyComponent.Config.BindAddr)
	uiFeature := componenttest.NewUIFeature(url)
	uiFeature.RegisterSteps(ctx)

	apiFeature.RegisterSteps(ctx)
	redirectProxyComponent.RegisterSteps(ctx)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		uiFeature.Reset()
		redirectProxyComponent.Reset()

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		uiFeature.Close()
		if closeErr := redirectProxyComponent.Close(); closeErr != nil {
			panic(closeErr)
		}

		return ctx, nil
	})
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

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
