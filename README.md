# Promsaint
[Nagios](https://www.nagios.org) -> [Prometheus Alertmanager](https://prometheus.io) integration

When transitioning from Nagios to Prometheus it is useful to be able to maintain the alerts configured in nagios. Promsaint allows you to push alerts generated in Nagios into the Prometheus Alertmanager so that you can have one consistent alerting and control plane.

## Names

"Prometheus Ain't Gonna Insist On Sainthood"

Promsaint is a portmanteau of Prometheus and the original name for Nagios [NetSaint](https://en.wikipedia.org/wiki/Nagios).

## Building

A Makefile is provided with this project to make it extremely easy to just clone and build all the binaries

    make

## Running the Daemon

    ./promsaint

    Usage of ./bin/promsaint:
      -alertmanager string
            Alertmanager host (default "http://localhost:9093")
      -forgetage duration
            How long to persist an alert after nagios forgets about it before removing it (default 1m0s)
      -generator_url string
            Breakdown Source for nagios (default "https://nagios.example.com/nagios/cgi-bin/status.cgi")
      -listen string
            HTTP endpoint (default ":8080")
      -log.format value
            Set log formatter
      -log.level value
            Set log level
      -pruneage duration
            How old of alerts to keep around (default 1m0s)
      -pruneinterval duration
            How often to prune the database (default 2m0s)
      -publishinterval duration
            How often to publish the database (default 1m0s)
      -publishminimum duration
            Minimum time to wait between alertmanager pushes (default 5s)

### Docker

Alternately you can start an Promsaint and Alertmanager instance automatically using `docker-compose`

    docker-compose up

## Running the CLI

The primary interaction method from Nagios to the Promsaint Daemon is done through the included promsaint-cli.

### Host Alerts

    ./promsaint-cli -hostalert -host testbox1 -state DOWN -ntype PROBLEM

And the recovery

    ./promsaint-cli -hostalert -host testbox1 -state UP -ntype RECOVERY

### Service Alert

    ./promsaint-cli -servicealert -host testbox1 -service exampleSVC -state CRITICAL -ntype PROBLEM -msg "Example Service is Down!" -note "Check example for a fix"

Recovery

    ./promsaint-cli -servicealert -host testbox1 -service exampleSVC -state OK -ntype RECOVERY

## Nagios Configuration

Included in this repo is a sample config file that use can use to start sending alerts from Nagios into Alertmanager.

[promsaint.cfg](nagios/promsaint.cfg)

Drop that file in your Nagios config directory and include the users in contact sections of your configured services and after a quick reload you should see alerts appearing in Alertmanager!

## Smoke Tests

An end-to-end testing tool is provided `promsaint-smoke` which will create a testing alert in promsaint and check alertmanager to ensure  that it was created there. At this point it will resolve the alert.

    ./promsaint-smoke

## Multiple Alertmanagers

I haven't writen this yet :p. Someday in the future I would like to get to this.

Pull Requests Welcome!
