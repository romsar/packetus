# Packetus

**Packetus** is a small, open-source program that shows the history of package changes in the project.

At the moment Packetus supports `composer.json` and `package.json` files.

## Install
Install _Packetus_ using `go install` command:
```shell
go install github.com/romsar/packetus@latest
```

## Usage
Run _Packetus_ using next command:
```shell
packetus <path to the repo> <path to the package manager file> --some-option=some-value
```

### Options
| Option          | Description                                                                          | Type                                       | Default value           |
|-----------------|--------------------------------------------------------------------------------------|--------------------------------------------|-------------------------|
| --commits       | Processed commits count                                                              | uint                                       | 100                     |
| --strategy      | Strategy name: composer, npm. If no specified, strategy will be chosen by file name. | string enum: composer, npm                 |                         |
| --change-events | Change events, that will be captured.                                                | string enum array: added, updated, deleted | added, updated, deleted |
| --dev           | Capture dev packages.                                                                | boolean                                    | true                    |
| --json          | Path to the file where results will be saved in json format.                         | string                                     |                         |
| --csv           | Path to the file where results will be saved in csv format.                          | string                                     |                         |

### Examples

Get changes for last 500 commits:
```shell
packetus . "/Users/<some user>/project" "./composer.json" --commits=500
```

Get only added and deleted changes:
```shell
packetus . "/Users/<some user>/project" "./composer.json" --change-events="added,deleted"
```

With specified strategy:
```shell
packetus . "/Users/<some user>/project" "./foo.json" --strategy=npm
```

Do not capture dev packages:
```shell
packetus . "/Users/<some user>/project" "./package.json" --dev=false
```

Save results in json format:
```shell
packetus . "/Users/<some user>/project" "./package.json" --json="/Users/<some user>/result.json"
```

Save results in csv format:
```shell
packetus . "/Users/<some user>/project" "./package.json" --csv="/Users/<some user>/result.csv"
```
