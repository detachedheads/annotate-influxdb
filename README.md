# annotate-influxdb [![Release][release-image]][release-url] [![Build Status][travis-image]][travis-url]

`annotate-influxdb` is a simple command line tool used to generate annotation data points for InfluxDB.

## Installation

A binary for various operating systems is available through [Github Releases][github-releases].  Download the appropriate archive, and extract into a directory within your PATH.

## Usage

For the full list of options:

```shell
$ annotate-influxdb --help
```

To see the version of `annotate-influxdb` you can use the following:

```shell
$ annotate-influxdb --version
```

## Getting Started

Once you have `annotate-influxdb` installed, you will need to know the address of your InfluxDB instance, as well as the name of the database and measurement where the annotation event will added.  

The following will create an event in the `annotation` database with a measurement named `deploy` with title and description describing the event. As no URL has been specified to an InfluxDB instance, the default of `http://localhost:8086` will be used.

```shell
$ annotate-influxdb --database annotation --measurement deploy --title "somecontainer:1.0.0" --description "somecontainer:1.0.0 has been deployed to development"
```

To help provide as much context as possible to an event you can add additional metadata in the form of [InfluxDB Tags][influxdb-tags].  These tags can then be used to filter and/or for display.  Building off of our previous command, this adds tags related to the container and environment.

```shell
$ annotate-influxdb --database annotation --measurement deploy --title "somecontainer:1.0.0" --description "somecontainer:1.0.0 has been deployed to development" --tag development --tag somecontainer
```

As you can see from the command you can add multiple tags by repeating the argument.

### Configuration File Format

`annotate-influxdb` provides the ability to configure all command line arguments within a configuration to allow for predefined annotations to be added without having to leverage the full compliment of command line arguments.

Configuration files can be written in YAML, TOML or JSON.

```yaml
# This block specifies the InfluxDB configuration
influxdb:
  # This specifies the URL of the InfluxDB instance
  url: 'http://internal.influxdb.service.consul:8086'
  # This is the databae where the measurement will be created.
  database: 'events'
  # This is the description of the annotation.
  description: 'This is a description for the annotation!'
  # This is the measurement for the annotation.
  measurement: 'events'
  # This is the title of the annotation.
  title: 'This is a title for the annotation!'
  # This is a tag for the annotation.
  tag:
    - 'tag1'
    - 'tag2'
```
[docker]: https://www.docker.com
[docker-compose]: https://docs.docker.com/compose/
[docker-golang]: https://hub.docker.com/_/golang/
[github-releases]: https://github.com/detachedheads/annotate-influxdb/releases
[go]: https://www.golang.org/
[influxdb-tags]: https://docs.influxdata.com/influxdb/v1.2/concepts/glossary/#tag
