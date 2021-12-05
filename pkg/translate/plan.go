package main

import (
	"context"
	"log"
	"os"
	"time"

	translate "cloud.google.com/go/translate/apiv3"
	"google.golang.org/api/option"
	translatepb "google.golang.org/genproto/googleapis/cloud/translate/v3"

	pkgT "github.com/jessicagreben/wide-load/pkg/tester"
)

type testPlan struct {
	test *pkgT.TestFramework
}

func (t testPlan) Execute(config pkgT.Config) {
	t.test = pkgT.NewTestFramework(config, &testScenario{test: t.test})
	t.test.Run()
	t.test.Results.Report()
}

func (t *testPlan) Stop() {
	// t.test.Stop()
}

type testScenario struct {
	test       *pkgT.TestFramework
	ctx        context.Context
	client     *translate.TranslationClient
	logVerbose bool
}

func (t *testScenario) SetupOnce() {
	log.Println("translate setup once executing")
	credsPath := os.Getenv("GCP_CREDS_PATH")
	if credsPath == "" {
		log.Fatalln("env var $GCP_CREDS_PATH needs to be set to path of file with GCP credentials")
	}
	ctx := context.Background()
	c, err := translate.NewTranslationClient(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		log.Fatalln("NewTranslationClient", err)
	}
	t.client = c
	t.ctx = ctx
	t.logVerbose = os.Getenv("DEBUG") != ""
}

func (t *testScenario) Setup() {
}

func (t *testScenario) Test() (int64, error) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatalln("env var $GCP_PROJECT_ID needs to be set")
	}

	req := &translatepb.TranslateTextRequest{
		Contents:           []string{"<h1>hello world!</h1>"},
		MimeType:           "text/html",
		SourceLanguageCode: "en",
		TargetLanguageCode: "es",
		Parent:             projectID,
	}
	start := time.Now()
	resp, err := t.client.TranslateText(t.ctx, req)
	if err != nil {
		log.Fatalln("TranslateText", err)
		return 0, err
	}
	latency := time.Since(start)
	log.Println("api latency:", latency)
	if t.logVerbose {
		for _, t := range resp.Translations {
			log.Println("text:", t.GetTranslatedText())
		}
	}
	return latency.Milliseconds(), nil
}

func (t *testScenario) Cleanup() {
	log.Println("translate cleanup executing")
	t.client.Close()
}

var (
	TestPlan     testPlan
	TestScenario testScenario
)
