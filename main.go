package main

import (
	"compress/bzip2"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	AbstractIndexDE            = "https://dumps.wikimedia.org/dewiki/latest/dewiki-latest-pages-articles.xml.bz2"
	PersonDataTemplateRegExpDE = regexp.MustCompile(`(?i:\{\{personendaten([^\}]+)\}\})`)
	TemplateFieldsRegExp       = regexp.MustCompile(`(?i:\s*([a-z]+)\s*=[\t\n\f\r '"ʿ]*(.+)[\t\n\f\r '"ʿ]*)`)
	NameSeperatorRegExp        = regexp.MustCompile(`\s*,\s*`)
	FirstNameSeperatorRegExp   = regexp.MustCompile(`[\t\n\f\r \-\.'"ʿ]`)
)

// Wikipedia XML
type WikipediaRevision struct {
	ID       int    `xml:"id"`
	ParentID int    `xml:"parentid"`
	Text     string `xml:"text"`
}

type WikipediaPage struct {
	Title     string               `xml:"title"`    // Title in text form. (Using spaces, not underscores; with namespace )
	Namespace string               `xml:"ns"`       // Namespace in canonical form
	ID        int                  `xml:"id"`       // Optional page ID number
	Redirect  string               `xml:"redirect"` // Flag if the current revision is a redirect
	Revision  []*WikipediaRevision `xml:"revision"` // Set of revisions
}

// Firstname + Count is used for sorting
type FirstnameCount struct {
	Firstname string // Firstname
	Count     int    // Count
}

func (m FirstnameCount) String() string { return fmt.Sprintf("[%d : %s]", m.Count, m.Firstname) }

type FirstnameCounts []*FirstnameCount

func (m FirstnameCounts) Len() int           { return len(m) }
func (m FirstnameCounts) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m FirstnameCounts) Less(i, j int) bool { return m[i].Count > m[j].Count }

// Main entry point
func main() {
	// Print banner
	color.NoColor = false

	color.HiCyan("                                          __ __      __    ")
	color.HiCyan(".-.--..---.-.--.-.--.-----.-----._____.--|  |__|----|  |_  ")
	color.HiCyan("|  .  |  -  |  . .  |  -__|__ --|_____|  -  |  |  --|   _| ")
	color.HiCyan("|__|__|___._|__|-|__|_____|_____|     |_____|__|____|_____|")
	color.HiCyan("                                                           ")

	// Cobra command
	cmd := &cobra.Command{
		Use:     "names-dict",
		Long:    "Create a password dictionary based on names.",
		Args:    cobra.NoArgs,
		Version: "0.0.1",
		Run:     namesDict,
	}

	cmd.Flags().BoolP("verbose", "v", false, "write more")

	cmd.Flags().StringP("dump-url", "u", "", "overwrite default URL for given language")
	cmd.Flags().IntP("count", "c", 0, "take the top N names only (0 means 'all')")

	// Viper config
	viper.SetEnvPrefix("NAMES_DICT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlags(cmd.Flags())

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/names-dict")
	viper.AddConfigPath("$HOME/.config/names-dict")
	viper.AddConfigPath(".")

	viper.ReadInConfig()

	// Run command
	cmd.Execute()
}

// aykroyd is called if the CLI interfaces has been satisfied.
func namesDict(cmd *cobra.Command, args []string) {
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

	// Decompress Bzip2
	decr := bzip2.NewReader(resp.Body)

	// Streamed XML parsing
	firstnameHist := make(map[string]int)

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
							firstname := FirstNameSeperatorRegExp.Split(name[1], -1)
							if len(firstname) < 1 {
								continue
							}

							// Increment usage
							firstnameHist[firstname[0]] += 1
						}
					}
				}
			}
		default:
		}
	}

	// Sort names
	final := make([]*FirstnameCount, 0, len(firstnameHist))

	for f, c := range firstnameHist {
		final = append(final, &FirstnameCount{
			Firstname: f,
			Count:     c,
		})
	}

	sort.Sort(FirstnameCounts(final))

	// Limit by given number
	if cnt := viper.GetInt("count"); cnt > 0 {
		final = final[0:cnt]
	}

	// ...
	fmt.Println(final)
}
