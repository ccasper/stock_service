# Stock Analysis Service
This provides very rudimentary evaluation of a stock ticker.

A golang stock service that can run on a headless server. This includes building, archiving, packaging (into deb packages), and running on a debian/ubuntu arm or amd server

This is an excellent fully running scaffolding / skeleton to build your own golang application on.

- Creates a debian package that run on both arm64 and amd64 (Works on Oracle ARM free-tier, Google Cloud, Vultr, etc)
  - It builds both binaries and packages them together with a run wrapper for selecting the architecture.
  - Service files are placed in `/opt/stock/`, and ticker caching (per hour) is stored in `/opt/stock/data`.
    - Some installed configurations, logrotate and systemd are installed in /etc
    - `apt-get --purge remove stock` will remove all files related to this package, including the data/cache and logs, as if the package was never installed.
- Creates a source archive that is within the deb package in `/opt/stock/src` for understanding what code is running on the server.
- Has versioning in a golang file that the build scripts use so the main file can also know it's version.
- Creates the necessary ufw, systemd, logrotate files to make the service automatically start and reload.
- Uses systemd watchdog to ensure the service doesn't get stuck.
- Uses a separate health check port for verifying installation was successful.
- The log files will be viewed at `tail -f /var/log/stock.log`
- The systemd log for stock can be read by: `journalctl -u stock` or tailed by adding `-f`
- This service runs on port 8080 for all interfaces by default. This can be configured in the build.sh file.
- The `build.sh` is located in the `cmd/stock` folder because the root hierarchy could be used by multiple golang services in your `cmd/proxy`,`cmd/stock`,`cmd/auth` system that each would likely have unique build.sh needs.
- For shared libraries across services, put them in `internal/<package>/<go files>`

# Screenshots of the service running:

![Analysis of stock valuation](docs/stock_analysis_MSFT.png)

![Entered stock and loading data](docs/stock_metrics_analyzer_home.png)

- Note: Loading can take a few seconds due to using free parsing of Yahoo pages, and requires running a headless chromium browser to gain proper access to these in-detail metrics for free.

# Instructions

## Run locally

go run ./cmd/stock

## Install on cloud/remote SSH machine

``` bash
SERVER=myfqdn or IP address

./cmd/stock/build.sh && scp -r cmd/stock/stock_1.0.0.deb ${SERVER?}: && 
ssh ${SERVER?} "sudo bash -c 'apt-get -y remove stock ; dpkg -i stock_1.0.0.deb && apt-get install -f'"
```

- We call `apt-get install -f` at the end to get any missing deb dependencies.

## If you want rollback safe dpkg installation:

- Copy tools/safe-dpkg on your service host server in /usr/local/bin

``` bash
SERVER=myfqdn or IP address

scp tools/safe-dpkg ${SERVER?}:
ssh ${SERVER?} "sudo bash -c 'chown root:root safe-dpkg && chmod 755 safe-dpkg && mv safe-dpkg /usr/local/bin/safe-dpkg'"
```

- Call `safe-dpkg stock_1.0.0.deb` instead of apt-get -y remove stock ; dpkg -i stock_1.0.0.deb. It will also safely roll back to the previous good running package on install/health failure (even if it has the same version name).

  - This will store a copy of the deb file in `/var/cache/safe-dpkg` only if it's successful in installation and healthy.
  - During install, it will remove the old package as needed.
  - On failure, it rolls back to the newest deb file in `/var/cache/safe-dpkg`.
  - On dependency issues, it will automatically install the dependencies.

``` bash
SERVER=myfqdn or IP address

./cmd/stock/build.sh && scp -r cmd/stock/stock_1.0.0.deb ${SERVER?}: && 
ssh ${SERVER?} "sudo bash -c 'apt-get -y remove stock ; safe-dpkg stock_1.0.0.deb || apt-get install -f'"
```

## ARM64 warning

Although all the tooling here works with ARM64, and the main binary and page will load, there are some unsolved technical issues with using chromedp (chromium/chrome headless) on ARM64 to fetch actual ticker data that have not been addressed.

The current error is:
`cmd_run.go:1400: WARNING: cannot create user data directory: cannot create snap home dir: mkdir`