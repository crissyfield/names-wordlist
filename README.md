<p align="center"><img width="200" src="etc/hero.svg"></a></p>

<p align="center">
  <a href="https://godoc.org/github.com/crissyfield/names-wordlist"><img src="https://godoc.org/github.com/crissyfield/names-wordlist?status.svg" alt="GoDoc"></a>
  <a href="https://goreportcard.com/report/github.com/crissyfield/names-wordlist"><img src="https://goreportcard.com/badge/github.com/crissyfield/names-wordlist" alt="Go Report Card"></a>
  <a href="http://opensource.org/licenses/MIT"><img src="http://img.shields.io/badge/license-MIT-brightgreen.svg" alt="MIT License"></a>
</p>


# Names Wordlist

**Names-Wordlist** is a command line tool that extracts popular first names from Wikipedia dumps to generate
wordlist for password cracking.


## Installation

### Wordlists

Pre-generated wordlists can be downloaded directly from the
[release page](https://github.com/crissyfield/names-wordlist/releases/latest).

### Binaries

Pre-built binaries are available from the
[release page](https://github.com/crissyfield/names-wordlist/releases/latest) as well. Simply download, make
executable, and move it to a folder in your `PATH`:

```bash
curl -L https://github.com/crissyfield/names-wordlist/releases/download/v1.0.0/names-wordlist-`uname -s`-`uname -m` >/tmp/names-wordlist
chmod +x /tmp/names-wordlist
sudo mv /tmp/names-wordlist /usr/local/bin/names-wordlist
```


## Usage

### Binary

Run `names-wordlist` in the command line like this:

```bash
names-wordlist output.lst
```


## License

Copyright (c) 2019 Crissy Field GmbH. Released under the
[MIT License](https://github.com/crissyfield/names-wordlist/blob/master/LICENSE).

Dictionary icon made by [Freepik](https://www.flaticon.com/authors/freepik) from
[www.flaticon.com](https://www.flaticon.com/).
