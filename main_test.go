package main

import (
	"context"
	"flag"
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
	component, err := steps.NewComponent()
	if err != nil {
		panic(err)
	}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		component.Reset()

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if closeErr := component.Close(); closeErr != nil {
			panic(closeErr)
		}

		return ctx, nil
	})

	component.RegisterSteps(ctx)
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
