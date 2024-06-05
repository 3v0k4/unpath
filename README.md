# Unpath

<div align="center">
  <img width="200" width="200" src=".github/images/unpath.svg" />
</div>

```bash
Usage: unpath UNCMD CMD

unpath runs CMD with a modified PATH that does not contain UNCMD.

Arguments:
  UNCMD the command to hide from PATH
  CMD   the command to run with the modified PATH

Examples:
  unpath cat ./script script-arg

  unpath cat CMD subcmd-arg

  unpath cat unpath env CMD
```

## Installation

You can install unpath with Go:

```bash
go install github.com/3v0k4/unpath
```

Or fetch the executable from GitHub:

```bash
# PLATFORM {linux,darwin}
# ARCHITECTURE {amd64,arm64}
curl https://github.com/3v0k4/unpath/releases/download/v0.1.0/unpath-PLATFORM-ARCH --output unpath
chmod +x unpath
./unpath
```

## Usage

```bash
unpath cat ./script script-arg

unpath cat command command-arg
```

To show all the options:

```bash
unpath
```

## Development

Unpath is dependency-free (it only uses the Go standard library), so there are no prerequisites.

```bash
go test
```

To release a new version add a tag and push:

```bash
git tag vX.Y.Z
git push --tags
```

## Contributing

Bug reports and pull requests are welcome on [GitHub](https://github.com/3v0k4/unpath).

## License

The module is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
