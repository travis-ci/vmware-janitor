package main

import (
	"time"

	"github.com/urfave/cli"
)

var (
	Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "u, vsphere-url",
			Usage:  "URL of the vsphere server, including '/sdk' if applicable",
			EnvVar: "VSPHERE_JANITOR_VSPHERE_URL,VSPHERE_URL",
		},
		cli.StringSliceFlag{
			Name:   "p, vsphere-vm-paths",
			Usage:  "**REQUIRED**: Paths in inventory that contain VMs for cleanup",
			EnvVar: "VSPHERE_JANITOR_VSPHERE_VM_PATHS,VSPHERE_VM_PATHS",
		},
		cli.BoolFlag{
			Name:   "S, skip-destroy",
			Usage:  "Do not destroy VMs -- only power down",
			EnvVar: "VSPHERE_JANITOR_SKIP_DESTROY,SKIP_DESTROY",
		},
		cli.BoolTFlag{
			Name:   "z, skip-zero-uptime",
			Usage:  "Skip over VMs with zero uptime",
			EnvVar: "VSPHERE_JANITOR_SKIP_ZERO_UPTIME,SKIP_ZERO_UPTIME",
		},
		cli.BoolTFlag{
			Name:   "B, skip-no-boot-time",
			Usage:  "Skip over VMs without a boot time",
			EnvVar: "VSPHERE_JANITOR_SKIP_NO_BOOT_TIME,SKIP_NO_BOOT_TIME",
		},
		cli.DurationFlag{
			Name:   "C, cutoff",
			Value:  2 * time.Hour,
			Usage:  "Max uptime cutoff",
			EnvVar: "VSPHERE_JANITOR_CUTOFF,CUTOFF",
		},
		cli.DurationFlag{
			Name:   "zero-uptime-cutoff",
			Value:  5 * time.Minute,
			Usage:  "Max 'zero uptime' cutoff",
			EnvVar: "VSPHERE_JANITOR_ZERO_UPTIME_CUTOFF,ZERO_UPTIME_CUTOFF",
		},
		cli.IntFlag{
			Name:   "c, concurrency",
			Usage:  "Concurrent cleanup goroutine count",
			EnvVar: "VSPHERE_JANITOR_CONCURRENCY,CONCURRENCY",
		},
		cli.BoolFlag{
			Name:   "O, once",
			Usage:  "Only run one cleanup",
			EnvVar: "VSPHERE_JANITOR_ONCE,ONCE",
		},
		cli.DurationFlag{
			Name:   "s, cleanup-loop-sleep",
			Value:  1 * time.Minute,
			Usage:  "Sleep interval between cleaning up all paths",
			EnvVar: "VSPHERE_JANITOR_CLEANUP_LOOP_SLEEP,CLEANUP_LOOP_SLEEP",
		},
		cli.IntFlag{
			Name:   "R, rate-per-second",
			Value:  5,
			Usage:  "Rate limit max vms handled per second",
			EnvVar: "VSPHERE_JANITOR_RATE_PER_SECOND,RATE_PER_SECOND",
		},
		cli.StringFlag{
			Name:   "librato-email",
			Usage:  "Librato metrics account email",
			EnvVar: "VSPHERE_JANITOR_LIBRATO_EMAIL,LIBRATO_EMAIL",
		},
		cli.StringFlag{
			Name:   "librato-token",
			Usage:  "Librato metrics account token",
			EnvVar: "VSPHERE_JANITOR_LIBRATO_TOKEN,LIBRATO_TOKEN",
		},
		cli.StringFlag{
			Name:   "librato-source",
			Usage:  "Librato metrics source name",
			EnvVar: "VSPHERE_JANITOR_LIBRATO_SOURCE,LIBRATO_SOURCE",
		},
		cli.BoolFlag{
			Name:   "silence-metrics",
			Usage:  "Disable logging metrics to stderr",
			EnvVar: "VSPHERE_JANITOR_SILENCE_METRICS,SILENCE_METRICS",
		},
		cli.StringFlag{
			Name:   "honeycomb-write-key",
			Usage:  "Honeycomb write key",
			EnvVar: "VSPHERE_JANITOR_HONEYCOMB_WRITE_KEY,HONEYCOMB_WRITE_KEY",
		},
		cli.StringFlag{
			Name:   "honeycomb-dataset",
			Usage:  "Honeycomb dataset name for cleanup events",
			EnvVar: "VSPHERE_JANITOR_HONEYCOMB_DATASET,HONEYCOMB_DATASET",
		},
		cli.StringFlag{
			Name:   "pprof-port",
			Usage:  "Port to set up net/http/pprof on",
			EnvVar: "VSPHERE_JANITOR_PPROF_PORT,PPROF_PORT",
		},
	}
)
