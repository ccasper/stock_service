module app

go 1.25

require (
	github.com/chromedp/cdproto v0.0.0-20250803210736-d308e07a266d
	github.com/chromedp/chromedp v0.14.2
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	maragu.dev/gomponents v1.0.0
)

require (
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/go-json-experiment/json v0.0.0-20250910080747-cc2cfa0554c3 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.4.0
