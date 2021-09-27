package main

import (
	"encoding/json"
	"os"

	plugin "github.com/drone-plugins/drone-github-actions"
	"github.com/drone-plugins/drone-github-actions/daemon"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "unknown"
)

type genericMapType struct {
	m      map[string]string
	strVal string
}

func (g *genericMapType) Set(value string) error {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(value), &m); err != nil {
		return err
	}
	g.m = m
	g.strVal = value
	return nil
}

func (g *genericMapType) String() string {
	return g.strVal
}

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		if err := godotenv.Load(env); err != nil {
			logrus.Fatal(err)
		}
	}

	app := cli.NewApp()
	app.Name = "drone github actions plugin"
	app.Usage = "drone github actions plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "action-name",
			Usage:  "Github action name",
			EnvVar: "PLUGIN_USES",
		},
		cli.StringFlag{
			Name:   "action-with",
			Usage:  "Github action with",
			EnvVar: "PLUGIN_WITH",
		},
		cli.StringFlag{
			Name:   "action-env",
			Usage:  "Github action env",
			EnvVar: "PLUGIN_ENV",
		},
		cli.BoolFlag{
			Name:   "action-verbose",
			Usage:  "Github action enable verbose logging",
			EnvVar: "PLUGIN_VERBOSE",
		},
		cli.StringFlag{
			Name:   "action-image",
			Usage:  "Image to use for running github actions",
			Value:  "node:12-buster-slim",
			EnvVar: "PLUGIN_ACTION_IMAGE",
		},

		// daemon flags
		cli.StringFlag{
			Name:   "docker.registry",
			Usage:  "docker daemon registry",
			Value:  "https://index.docker.io/v1/",
			EnvVar: "PLUGIN_DAEMON_REGISTRY",
		},
		cli.StringFlag{
			Name:   "daemon.mirror",
			Usage:  "docker daemon registry mirror",
			EnvVar: "PLUGIN_DAEMON_MIRROR",
		},
		cli.StringFlag{
			Name:   "daemon.storage-driver",
			Usage:  "docker daemon storage driver",
			EnvVar: "PLUGIN_DAEMON_STORAGE_DRIVER",
		},
		cli.StringFlag{
			Name:   "daemon.storage-path",
			Usage:  "docker daemon storage path",
			Value:  "/var/lib/docker",
			EnvVar: "PLUGIN_DAEMON_STORAGE_PATH",
		},
		cli.StringFlag{
			Name:   "daemon.bip",
			Usage:  "docker daemon bride ip address",
			EnvVar: "PLUGIN_DAEMON_BIP",
		},
		cli.StringFlag{
			Name:   "daemon.mtu",
			Usage:  "docker daemon custom mtu setting",
			EnvVar: "PLUGIN_DAEMON_MTU",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns",
			Usage:  "docker daemon dns server",
			EnvVar: "PLUGIN_DAEMON_CUSTOM_DNS",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns-search",
			Usage:  "docker daemon dns search domains",
			EnvVar: "PLUGIN_DAEMON_CUSTOM_DNS_SEARCH",
		},
		cli.BoolFlag{
			Name:   "daemon.insecure",
			Usage:  "docker daemon allows insecure registries",
			EnvVar: "PLUGIN_DAEMON_INSECURE",
		},
		cli.BoolFlag{
			Name:   "daemon.ipv6",
			Usage:  "docker daemon IPv6 networking",
			EnvVar: "PLUGIN_DAEMON_IPV6",
		},
		cli.BoolFlag{
			Name:   "daemon.experimental",
			Usage:  "docker daemon Experimental mode",
			EnvVar: "PLUGIN_DAEMON_EXPERIMENTAL",
		},
		cli.BoolFlag{
			Name:   "daemon.debug",
			Usage:  "docker daemon executes in debug mode",
			EnvVar: "PLUGIN_DAEMON_DEBUG",
		},
		cli.BoolFlag{
			Name:   "daemon.off",
			Usage:  "don't start the docker daemon",
			EnvVar: "PLUGIN_DAEMON_OFF",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.String("action-name") == "" {
		return errors.New("uses attribute must be set")
	}

	actionWith, err := strToMap(c.String("action-with"))
	if err != nil {
		return errors.Wrap(err, "with attribute is not of map type with key & value as string")
	}
	actionEnv, err := strToMap(c.String("action-env"))
	if err != nil {
		return errors.Wrap(err, "env attribute is not of map type with key & value as string")
	}

	plugin := plugin.Plugin{
		Action: plugin.Action{
			Uses:    c.String("action-name"),
			With:    actionWith,
			Env:     actionEnv,
			Verbose: c.Bool("action-verbose"),
			Image:   c.String("action-image"),
		},
		Daemon: daemon.Daemon{
			Registry:      c.String("docker.registry"),
			Mirror:        c.String("daemon.mirror"),
			StorageDriver: c.String("daemon.storage-driver"),
			StoragePath:   c.String("daemon.storage-path"),
			Insecure:      c.Bool("daemon.insecure"),
			Disabled:      c.Bool("daemon.off"),
			IPv6:          c.Bool("daemon.ipv6"),
			Debug:         c.Bool("daemon.debug"),
			Bip:           c.String("daemon.bip"),
			DNS:           c.StringSlice("daemon.dns"),
			DNSSearch:     c.StringSlice("daemon.dns-search"),
			MTU:           c.String("daemon.mtu"),
			Experimental:  c.Bool("daemon.experimental"),
		},
	}
	return plugin.Exec()
}

func strToMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	if s == "" {
		return m, nil
	}

	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return m, nil
}
