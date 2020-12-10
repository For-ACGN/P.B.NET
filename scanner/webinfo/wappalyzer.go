package webinfo

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// WappalyzerURL is the link to the latest wappalyzer technologies.json file URL.
const WappalyzerURL = "https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/technologies.json"

// TechDef type encapsulates the json encoding of the whole technologies.json file.
type TechDef struct {
	Cats  map[string]*Category `json:"categories"`
	Techs map[string]*Tech     `json:"technologies"`
}

// Category names defined by wappalyzer.
type Category struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// Tech type encapsulates all the data about an Tech from technologies.json
type Tech struct {
	// raw data
	Cats        StringArray            `json:"cats"`
	Description string                 `json:"description"`
	URL         StringArray            `json:"url"`
	Headers     map[string]string      `json:"headers"`
	Cookies     map[string]string      `json:"cookies"`
	HTML        StringArray            `json:"html"`
	Meta        map[string]StringArray `json:"meta"`
	Script      StringArray            `json:"script"`
	Website     string                 `json:"website"`
	Implies     StringArray            `json:"implies"`

	// processed
	CatNames     []string      `json:"category_names"`
	URLRegex     []*TechRegexp `json:"-"`
	HeadersRegex []*TechRegexp `json:"-"`
	CookiesRegex []*TechRegexp `json:"-"`
	HTMLRegex    []*TechRegexp `json:"-"`
	MetaRegex    []*TechRegexp `json:"-"`
	ScriptRegex  []*TechRegexp `json:"-"`
}

// StringArray type is a wrapper for []string for use in unmarshalling
// the technologies.json.
type StringArray []string

// UnmarshalJSON is a custom unmarshaler for handling bogus technologies.json
// types from wappalyzer.
func (t *StringArray) UnmarshalJSON(data []byte) error {
	var sa []string
	err := json.Unmarshal(data, &sa)
	if err == nil {
		*t = sa
		return nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err == nil {
		*t = StringArray{s}
		return nil
	}
	var na []int
	err = json.Unmarshal(data, &na)
	if err == nil {
		sa := make(StringArray, len(na))
		for i := 0; i < len(na); i++ {
			sa[i] = strconv.Itoa(na[i])
		}
		*t = sa
		return nil
	}
	return errors.Errorf("failed to unmarshal json %s", data)
}

// TechRegexp is the technology regexp.
type TechRegexp struct {
	Regexp  *regexp.Regexp
	Name    string
	Version string
}

// LoadWappalyzerTechFile is used to load and process technologies.json.
func LoadWappalyzerTechFile(data []byte) (*TechDef, error) {
	td := new(TechDef)
	err := json.Unmarshal(data, td)
	if err != nil {
		return nil, err
	}
	for _, tech := range td.Techs {
		// set cat names
		catNames := make(StringArray, 0, len(tech.Cats))
		for i := 0; i < len(tech.Cats); i++ {
			cat, ok := td.Cats[tech.Cats[i]]
			if ok && cat.Name != "" {
				catNames = append(catNames, cat.Name)
			}
		}
		tech.CatNames = catNames
		// set regex with version
		tech.URLRegex, err = compileRegexes(tech.URL)
		if err != nil {
			return nil, err
		}
		tech.HTMLRegex, err = compileRegexes(tech.HTML)
		if err != nil {
			return nil, err
		}
		tech.ScriptRegex, err = compileRegexes(tech.Script)
		if err != nil {
			return nil, err
		}
	}
	return td, nil
}

func compileRegexes(s StringArray) ([]*TechRegexp, error) {
	var list []*TechRegexp
	for _, regexStr := range s {
		// split version detection
		split := strings.Split(regexStr, "\\;")
		regex, err := regexp.Compile(split[0])
		if err != nil {
			return nil, errors.Errorf("failed to compile %s", regexStr)
		}
		rv := TechRegexp{
			Regexp: regex,
		}
		if len(split) > 1 && strings.HasPrefix(split[1], "version:") {
			rv.Version = split[1][len("version:"):]
		}
		list = append(list, &rv)
	}
	return list, nil
}
