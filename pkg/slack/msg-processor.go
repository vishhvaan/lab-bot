package slack

import (
	"strings"
)

func (sc *slackClient) textMatcher(message string) (match string, err string) {
	message = strings.ToLower(message)
	match = ""
	err = "no match found"
	for m := range sc.responses {
		if strings.Contains(message, m) {
			if match == "" {
				match = m
				err = ""
			} else {
				return "", "multiple matches found"
			}
		}
	}
	return match, err
}

func OnOffDetector(message string) (detected string) {
	lm := strings.ToLower(message)
	if strings.Contains(lm, " on") && !strings.Contains(lm, " off") {
		return "on"
	} else if strings.Contains(lm, " off") && !strings.Contains(lm, " on") {
		return "off"
	} else {
		return "both"
	}
}
