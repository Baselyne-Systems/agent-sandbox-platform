package governance

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidInput = errors.New("invalid input")

// DataClassification represents the sensitivity level of data.
type DataClassification int

const (
	ClassificationPublic DataClassification = iota + 1
	ClassificationInternal
	ClassificationConfidential
	ClassificationRestricted
)

type pattern struct {
	name           string
	re             *regexp.Regexp
	classification DataClassification
}

// Service implements stateless data governance logic.
type Service struct {
	patterns             []pattern
	approvedDestinations map[string]bool
}

func NewService() *Service {
	return &Service{
		patterns: []pattern{
			{name: "ssn", re: regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), classification: ClassificationRestricted},
			{name: "credit_card", re: regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`), classification: ClassificationRestricted},
			{name: "aws_key", re: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), classification: ClassificationConfidential},
			{name: "email", re: regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`), classification: ClassificationInternal},
			{name: "phone", re: regexp.MustCompile(`\b\+?1?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`), classification: ClassificationInternal},
		},
		approvedDestinations: map[string]bool{
			"internal-api":   true,
			"secure-storage": true,
			"audit-log":      true,
		},
	}
}

// ClassifyData scans content for sensitive patterns and returns the highest
// classification level along with the names of all detected patterns.
func (s *Service) ClassifyData(content []byte, contentType string) (DataClassification, []string, error) {
	if len(content) == 0 {
		return 0, nil, ErrInvalidInput
	}

	text := string(content)
	var highest DataClassification
	var detected []string

	for _, p := range s.patterns {
		if p.re.MatchString(text) {
			detected = append(detected, p.name)
			if p.classification > highest {
				highest = p.classification
			}
		}
	}

	if highest == 0 {
		highest = ClassificationPublic
	}

	return highest, detected, nil
}

// InspectEgress combines classify + check in a single call. It classifies the
// content, checks the policy, and returns the combined result.
func (s *Service) InspectEgress(agentID, destination string, content []byte, contentType string) (bool, string, DataClassification, []string, error) {
	if agentID == "" || destination == "" {
		return false, "", 0, nil, ErrInvalidInput
	}

	classification, patterns, err := s.ClassifyData(content, contentType)
	if err != nil {
		return false, "", 0, nil, err
	}

	allowed, reason, err := s.CheckPolicy(agentID, destination, classification)
	if err != nil {
		return false, "", 0, nil, err
	}

	return allowed, reason, classification, patterns, nil
}

// CheckPolicy determines whether an agent is allowed to send data of a given
// classification to a destination.
func (s *Service) CheckPolicy(agentID, destination string, classification DataClassification) (bool, string, error) {
	if agentID == "" || destination == "" {
		return false, "", ErrInvalidInput
	}

	dest := strings.ToLower(destination)

	switch classification {
	case ClassificationRestricted:
		return false, "restricted data cannot be sent to any destination", nil
	case ClassificationConfidential:
		if s.approvedDestinations[dest] {
			return true, "confidential data allowed to approved destination", nil
		}
		return false, "confidential data denied to unapproved destination", nil
	default:
		return true, "data transfer allowed", nil
	}
}
