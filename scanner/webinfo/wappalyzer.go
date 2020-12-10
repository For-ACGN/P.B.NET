package webinfo

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// WappalyzerURL is the link to the latest wappalyzer technologies.json file URL.
const WappalyzerURL = "https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/technologies.json"

// AppsDefinition type encapsulates the json encoding of the whole technologies.json file.
type AppsDefinition struct {
	Cats  map[string]Category `json:"categories"`
	Techs map[string]App      `json:"technologies"`
}

// Category names defined by wappalyzer.
type Category struct {
	Name string `json:"name"`
}

// App type encapsulates all the data about an App from technologies.json
type App struct {
	// raw data
	Cats    StringArray            `json:"cats"`
	Cookies map[string]string      `json:"cookies"`
	Headers map[string]string      `json:"headers"`
	Meta    map[string]StringArray `json:"meta"`
	HTML    StringArray            `json:"html"`
	Script  StringArray            `json:"script"`
	URL     StringArray            `json:"url"`
	Website string                 `json:"website"`
	Implies StringArray            `json:"implies"`

	// after processed
	CatNames    []string    `json:"category_names"`
	HTMLRegex   []AppRegexp `json:"-"`
	ScriptRegex []AppRegexp `json:"-"`
	URLRegex    []AppRegexp `json:"-"`
	HeaderRegex []AppRegexp `json:"-"`
	MetaRegex   []AppRegexp `json:"-"`
	CookieRegex []AppRegexp `json:"-"`
}

// StringArray type is a wrapper for []string for use in unmarshalling the technologies.json.
type StringArray []string

// UnmarshalJSON is a custom unmarshaler for handling bogus technologies.json types from wappalyzer
func (t *StringArray) UnmarshalJSON(data []byte) error {
	var s string
	var sa []string
	var na []int

	if err := json.Unmarshal(data, &s); err != nil {
		if err := json.Unmarshal(data, &na); err == nil {
			// not a string, so maybe []int?
			*t = make(StringArray, len(na))

			for i, number := range na {
				(*t)[i] = fmt.Sprintf("%d", number)
			}

			return nil
		} else if err := json.Unmarshal(data, &sa); err == nil {
			// not a string, so maybe []string?
			*t = sa
			return nil
		}
		fmt.Println(string(data))
		return err
	}
	*t = StringArray{s}
	return nil
}

// AppRegexp is the app regexp.
type AppRegexp struct {
	Name    string
	Regexp  *regexp.Regexp
	Version string
}

// LoadWappalyzerTechFile is used to load and process technologies.json.
func LoadWappalyzerTechFile(data []byte) {

}
