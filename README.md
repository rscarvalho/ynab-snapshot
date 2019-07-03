# ynab-snapshot ![Build status](https://travis-ci.org/rscarvalho/ynab-snapshot.svg?branch=master)

A simple utility to take snapshots of a budget using the [You Need a Budget api](https://api.youneedabudget.com/).
It creates a CSV file with a snapshot of your categories in the specified path.

## Installation

```sh
$ go get github.com/rscarvalho/ynab-snapshot
```

## Usage

```sh
$ ynab-snapshot -h
Usage of ynab-snapshot:
  -Date string
        The month to take the snapshot in format YYYY-MM. (default "current")
  -Path string
        The target path for the snapshot. (default "/current/running/path")
  -Token string
        The ynab API token. Can be set by the environment variable $YNAB_TOKEN
```