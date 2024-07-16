package source_code

import "regexp"

var regexChallange = []*regexp.Regexp{
	regexp.MustCompile(`<title>.*Just a moment.*</title>`),
	regexp.MustCompile(`id="challenge-error-title"`),
	regexp.MustCompile(`id="challenge-error-text"`),
	regexp.MustCompile(`id="challenge-error"`),
	regexp.MustCompile(`challenge-platform`),
}

func HasChallange(data []byte) bool {
	for _, rc := range regexChallange {
		if rc.Match(data) {
			return true
		}
	}
	return false
}
