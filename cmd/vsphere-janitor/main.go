package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/libhoney-go"
	librato "github.com/mihasya/go-metrics-librato"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/travis-ci/vsphere-janitor"
	"github.com/travis-ci/vsphere-janitor/log"
	"github.com/travis-ci/vsphere-janitor/vsphere"
	"github.com/urfave/cli"
)

var (
	// VersionString is the git describe version set at build time
	VersionString = "?"
	// RevisionString is the git revision set at build time
	RevisionString = "?"
	// GeneratedString is the build date set at build time
	GeneratedString = "?"
)

func init() {
	cli.VersionPrinter = customVersionPrinter
	os.Setenv("VERSION", VersionString)
	os.Setenv("REVISION", RevisionString)
	os.Setenv("GENERATED", GeneratedString)
}

func customVersionPrinter(c *cli.Context) {
	fmt.Printf("%v v=%v rev=%v d=%v\n", c.App.Name, VersionString, RevisionString, GeneratedString)
}

func main() {
	app := cli.NewApp()
	app.Usage = "VMware vSphere cleanup thingy"
	app.Version = VersionString
	app.Author = "Travis CI GmbH"
	app.Email = "contact+vsphere-janitor@travis-ci.org"

	app.Flags = Flags
	app.Action = mainAction

	app.Run(os.Args)
}

func mainAction(c *cli.Context) error {
	ctx := context.Background()

	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	log.WithContext(ctx).Info("starting vsphere-janitor")
	defer func() { log.WithContext(ctx).Info("stopping vsphere-janitor") }()

	if c.String("pprof-port") != "" {
		go func() {
			log.WithContext(ctx).WithField("port", c.String("pprof-port")).Info("setting up pprof")
			http.ListenAndServe("localhost:"+c.String("pprof-port"), nil)
		}()
	}

	u, err := url.Parse(c.String("vsphere-url"))
	if err != nil {
		log.WithContext(ctx).WithError(err).Fatal("couldn't parse vSphere URL")
	}

	paths := c.StringSlice("vsphere-vm-paths")
	if len(paths) == 0 {
		log.WithContext(ctx).Fatal("missing vsphere vm paths")
	}

	cleanupLoopSleep := c.Duration("cleanup-loop-sleep")

	vSphereLister, err := vsphere.NewClient(ctx, u, true)
	if err != nil {
		log.WithContext(ctx).WithError(err).Fatal("couldn't create vsphere vm lister")
	}

	janitor := vspherejanitor.NewJanitor(vSphereLister, &vspherejanitor.JanitorOpts{
		Cutoff:           c.Duration("cutoff"),
		ZeroUptimeCutoff: c.Duration("zero-uptime-cutoff"),
		SkipDestroy:      c.Bool("skip-destroy"),
		Concurrency:      c.Int("concurrency"),
		RatePerSecond:    c.Int("rate-per-second"),
	})

	if c.String("librato-email") != "" && c.String("librato-token") != "" && c.String("librato-source") != "" {
		log.WithContext(ctx).Info("starting librato metrics reporter")

		go librato.Librato(metrics.DefaultRegistry, time.Minute,
			c.String("librato-email"), c.String("librato-token"), c.String("librato-source"),
			[]float64{0.95}, time.Millisecond)

		if !c.Bool("silence-metrics") {
			go metrics.Log(metrics.DefaultRegistry, time.Minute,
				log.WithContext(ctx).WithField("component", "metrics"))
		}
	}

	if c.String("honeycomb-write-key") != "" && c.String("honeycomb-dataset") != "" {
		log.WithContext(ctx).Info("configuring honeycomb reporting")

		libhoney.Init(libhoney.Config{
			WriteKey: c.String("honeycomb-write-key"),
			Dataset:  c.String("honeycomb-dataset"),
		})
		defer libhoney.Close()

		libhoney.AddDynamicField("meta.goroutines", func() interface{} { return runtime.NumGoroutine() })
		libhoney.AddField("app.version", c.App.Version)
		libhoney.AddField("service_name", c.String("librato-source"))
	}

	for {
		for _, path := range paths {
			err := janitor.Cleanup(ctx, path, time.Now())
			if err != nil {
				log.WithContext(ctx).WithError(err).Error("error cleaning up")
			}
		}

		if c.Bool("once") {
			log.WithContext(ctx).Info("finishing after one run")
			break
		}

		log.WithContext(ctx).WithField("duration", cleanupLoopSleep).Info("sleeping")
		time.Sleep(cleanupLoopSleep)
	}

	return nil
}
