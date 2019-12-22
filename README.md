<p align="center"><img width="200" src="etc/hero.svg"></a></p>

<p align="center">
  <a href="https://godoc.org/github.com/crissyfield/names-wordlist"><img src="https://godoc.org/github.com/crissyfield/names-wordlist?status.svg" alt="GoDoc"></a>
  <a href="https://goreportcard.com/report/github.com/crissyfield/names-wordlist"><img src="https://goreportcard.com/badge/github.com/crissyfield/names-wordlist" alt="Go Report Card"></a>
  <a href="http://opensource.org/licenses/MIT"><img src="http://img.shields.io/badge/license-MIT-brightgreen.svg" alt="MIT License"></a>
</p>


# Names-Wordlist

`names-wordlist` is a command line tool that extracts popular first names from Wikipedia dumps to generate a
wordlist for password cracking.

The wordlist addresses the pattern `{first name}{name}{special character}` that is commonly found in passwords.
The list has been highly effective in cracking passwords over the last few years.

To generate a wordlist, `names-wordlist` downloads the latest dump of the German Wikipedia and parses it for
person data (i.e. the `{{Personendaten}}` template). For each match, the first name is extracted and added to a
histogram. When a first name has occured more than `N` times (where `N` is configurable, but defaults to `1`)
it is stored in the wordlist.

In addition, each first name is written in lower-, upper-, and title-case and is appended by digits (of up to
`4` digits by default) and special characters (`!`, `$`, `@`, and `_` by default).


## :package: Installation

### Binaries

Pre-built binaries are available from the
[release page](https://github.com/crissyfield/names-wordlist/releases/latest). Simply download, make
executable, and move it to a folder in your `PATH`:

```bash
curl -L https://github.com/crissyfield/names-wordlist/releases/download/v1.0.0/names-wordlist-`uname -s`-`uname -m` > /tmp/names-wordlist
chmod +x /tmp/names-wordlist
sudo mv /tmp/names-wordlist /usr/local/bin/names-wordlist
```

### Wordlists

Pre-generated wordlists can be downloaded directly from the
[release page](https://github.com/crissyfield/names-wordlist/releases/latest) as well.

| Wordlist                                | Name Count |  Digits | Special Characters | Uncompressed |
|-----------------------------------------|-----------:|--------:|:------------------:|-------------:|
| `names-de-1count-4digits-special.txt`   |     45,534 | Up to 4 | `!`, `$`, `@`, `_` |    85.15 GiB |
| `names-de-2count-4digits-special.txt`   |     18,137 | Up to 4 | `!`, `$`, `@`, `_` |    33.11 GiB |
| `names-de-4count-4digits-special.txt`   |     19,078 | Up to 4 | `!`, `$`, `@`, `_` |    18.26 GiB |
| `names-de-8count-4digits-special.txt`   |      6,004 | Up to 4 | `!`, `$`, `@`, `_` |    10.87 GiB |
| `names-de-16count-4digits-special.txt`  |      3,649 | Up to 4 | `!`, `$`, `@`, `_` |     6.59 GiB |
| `names-de-32count-4digits-special.txt`  |      2,219 | Up to 4 | `!`, `$`, `@`, `_` |     4.00 GiB |

:speech_balloon: **Note:** The biggest dictionary takes roughly <u>6 minutes</u> on an NVIDIA GeForce 1080Ti to
crack NTLMv2 hashes.


## :computer: Usage

### Generate a Wordlist

Run `names-wordlist` in the command line like this:

```bash
names-wordlist output.lst
```

### Using the Wordlists

For instance, with [Hashcat](https://hashcat.net/hashcat/):

```bash
# Crack some NTLMv2 hashes
./hashcat64.bin -O --hash-type=5600 --attack-mode=0 hashes.txt names-de-1count-4digits-special.txt
```

## License

Copyright (c) 2019 Crissy Field GmbH. Released under the
[MIT License](https://github.com/crissyfield/names-wordlist/blob/master/LICENSE).

Dictionary icon made by [Freepik](https://www.flaticon.com/authors/freepik) from
[www.flaticon.com](https://www.flaticon.com/).
