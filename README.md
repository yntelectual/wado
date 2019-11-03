# WADO - watch and do

Wado is a simple utility take can watch a set for filesystem paths, and execute specified action when something happens.

## Installation

Via go get: `go get github.com/yntelectual/wado`

Manually download binary for your os from release page []()

## Usage

Wado needs a config file to do anything usefull. By default it looks for a file called `wado.json` in the current dir. You can specify a custom config file by specyfing `-c` param. 