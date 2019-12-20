package main

import (
	"compress/bzip2"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

const (
	AbstractIndexDE   = "https://dumps.wikimedia.org/dewiki/latest/dewiki-latest-pages-articles.xml.bz2"
	SpecialCharacters = "!$@_"
)

var (
	PersonDataTemplateRegExpDE = regexp.MustCompile(`(?i:\{\{personendaten([^\}]+)\}\})`)
	TemplateFieldsRegExp       = regexp.MustCompile(`(?i:\s*([a-z]+)\s*=[\t\n\f\r '"ʿ]*(.+)[\t\n\f\r '"ʿ]*)`)
	NameSeperatorRegExp        = regexp.MustCompile(`\s*,\s*`)
	FirstnameSeperatorRegExp   = regexp.MustCompile(`[\t\n\f\r \-\.'"ʿ]`)
)

// ...
type ProgressReader struct {
	bar    *mpb.Bar  // Progress bar
	reader io.Reader // Source reader
	prev   time.Time // Last time
}

func NewProgressReader(b *mpb.Bar, r io.Reader) *ProgressReader {
	return &ProgressReader{
		bar:    b,
		reader: r,
		prev:   time.Now(),
	}
}

func (m *ProgressReader) Read(p []byte) (int, error) {
	n, err := m.reader.Read(p)

	next := time.Now()
	m.bar.IncrInt64(int64(n), next.Sub(m.prev))
	m.prev = next

	return n, err
}

// Wikipedia XML
type WikipediaRevision struct {
	ID       int    `xml:"id"`
	ParentID int    `xml:"parentid"`
	Text     string `xml:"text"`
}

type WikipediaPage struct {
	Title     string               `xml:"title"`    // Title in text form. (Using spaces, not underscores; with namespace)
	Namespace string               `xml:"ns"`       // Namespace in canonical form
	ID        int                  `xml:"id"`       // Optional page ID number
	Redirect  string               `xml:"redirect"` // Flag if the current revision is a redirect
	Revision  []*WikipediaRevision `xml:"revision"` // Set of revisions
}

// Main entry point
func main() {
	// Print banner
	logoClr := color.New(color.FgHiCyan)

	logoClr.Fprintln(os.Stderr, "                                                              __ __ __       __    ")
	logoClr.Fprintln(os.Stderr, ".-.--..---.-.--.-.--.-----.-----._____.--._.--.-----.--.--.--|  |  |__|-----|  |_  ")
	logoClr.Fprintln(os.Stderr, "|  .  |  -  |  . .  |  -__|__ --|_____|  | |  |  -  |  .__|  -  |  |  |__ --|   _| ")
	logoClr.Fprintln(os.Stderr, "|__|__|___._|__|-|__|_____|_____|     |___.___|_____|__|  |_____|__|__|_____|_____|")
	logoClr.Fprintln(os.Stderr, "                                                                                   ")

	// Cobra command
	cmd := &cobra.Command{
		Use:     "names-wordlist",
		Long:    "Create wordlists based on Wikipedia person data.",
		Args:    cobra.ExactArgs(1),
		Version: "1.0.0",
		Run:     namesWordlist,
	}

	cmd.Flags().BoolP("verbose", "v", false, "write more")

	cmd.Flags().StringP("dump-url", "u", "", "overwrite default URL for given language")
	cmd.Flags().IntP("count", "c", 1, "ignore names with less than N occurences")
	cmd.Flags().IntP("digits", "d", 4, "append up to N digits after the name")
	cmd.Flags().StringP("special-chars", "s", SpecialCharacters, "append special characters from this set")

	// Viper config
	viper.SetEnvPrefix("NAMES_WORDLIST")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlags(cmd.Flags())

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/names-wordlist")
	viper.AddConfigPath("$HOME/.config/names-wordlist")
	viper.AddConfigPath(".")

	viper.ReadInConfig()

	// Run command
	cmd.Execute()
}

// aykroyd is called if the CLI interfaces has been satisfied.
func namesWordlist(cmd *cobra.Command, args []string) {
	// Set logging level
	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Download Wikipedia Dump
	dumpUrl := viper.GetString("dump-url")
	if dumpUrl == "" {
		dumpUrl = AbstractIndexDE
	}

	resp, err := http.Get(dumpUrl)
	if err != nil {
		logrus.Errorf("Unable to fetch abstract index: %w", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	// Show progress
	p := mpb.New()

	bar := p.AddBar(resp.ContentLength,
		mpb.PrependDecorators(decor.CountersKibiByte("% .2f / % .2f")),
		mpb.AppendDecorators(
			decor.Percentage(),
			decor.Name(" | ETA: "),
			decor.EwmaETA(decor.ET_STYLE_HHMMSS, 64),
		),
	)

	pr := NewProgressReader(bar, resp.Body)

	// Decompress Bzip2
	decr := bzip2.NewReader(pr)

	// Open output file
	out, err := os.Create(args[0])
	if err != nil {
		fmt.Errorf("Unable to create output file: %w", err)
		os.Exit(1)
	}

	defer out.Close()

	// Spin off output routne
	ch := make(chan string, 100)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go OutputRoutine(
		out,
		viper.GetInt("digits"),
		viper.GetString("special-chars"),
		ch,
		wg,
	)

	// Streamed XML parsing
	firstnameHist := make(map[string]int)
	cnt := viper.GetInt("count")

	decoder := xml.NewDecoder(decr)
	for {
		token, err := decoder.Token()
		if token == nil || err == io.EOF {
			break
		} else if err != nil {
			logrus.Errorf("Error decoding XML token: %w", err)
			os.Exit(1)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "page" {
				// Decode <page> element
				var p WikipediaPage

				if err = decoder.DecodeElement(&p, &t); err != nil {
					continue
				}

				// Skip if no or empty revision
				if len(p.Revision) == 0 || p.Revision[0] == nil {
					continue
				}

				// Iterate through all {{Persondata}} templates
				templates := PersonDataTemplateRegExpDE.FindAllStringSubmatch(p.Revision[0].Text, -1)
				for _, tmpl := range templates {
					// Split into fields
					for _, sub := range strings.Split(tmpl[1], "|") {
						// Parse key/value of field
						kv := TemplateFieldsRegExp.FindStringSubmatch(sub)
						if kv == nil {
							continue
						}

						switch strings.ToLower(kv[1]) {
						case "name":
							// Split last- and firstname
							name := NameSeperatorRegExp.Split(kv[2], -1)
							if len(name) < 2 {
								continue
							}

							// Split multiple firstnames
							firstname := FirstnameSeperatorRegExp.Split(name[1], -1)
							if len(firstname) < 1 {
								continue
							}

							// Increment usage
							firstnameHist[firstname[0]] += 1

							// Output
							if firstnameHist[firstname[0]] == cnt {
								ch <- firstname[0]
							}
						}
					}
				}
			}
		default:
		}
	}

	// Clean up output go routine
	close(ch)
	wg.Wait()
}

// ...
func OutputRoutine(w io.StringWriter, digits int, specialChars string, ch chan string, wg *sync.WaitGroup) {
	wg.Done()

	// Create number combinations
	digitCombs := []string{""}

	maxNumber := 1
	for d := 0; d < digits; d++ {
		maxNumber *= 10
		format := fmt.Sprintf("%%0%dd", d+1)

		for i := 0; i < maxNumber; i++ {
			digitCombs = append(digitCombs, fmt.Sprintf(format, i))
		}
	}

	// Create special character combinations
	charCombs := []string{""}

	for _, c := range specialChars {
		charCombs = append(charCombs, string(c))
	}

	// Generate output
	for name := range ch {
		// Lower case
		lwr := strings.ToLower(name)
		upr := strings.ToUpper(name)
		ttl := strings.Title(name)

		for _, d := range digitCombs {
			for _, c := range charCombs {
				w.WriteString(lwr + d + c + "\n" + upr + d + c + "\n" + ttl + d + c + "\n")
			}
		}
	}
}
