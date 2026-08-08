package main

import (
	"regexp"
)

var (
	// Redact AWS access key ID, secret access key, Bearer tokens, private keys, generic API tokens, and password flags
	awsAccessKeyRegex = regexp.MustCompile(`\b(AKIA|ASIA)[A-Z0-9]{16}\b`)
	bearerTokenRegex  = regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9\-\._~\+\/]+=*`)
	genericTokenRegex = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|passwd|auth)\s*[:=]\s*['"]?([^\s'"]+)['"]?`)
	flagPasswordRegex = regexp.MustCompile(`(?i)(--password|--token|--secret|-p)\s+['"]?([^\s'"]+)['"]?`)
	privateKeyRegex   = regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----[\s\S]*?-----END [A-Z ]+ PRIVATE KEY-----`)
)

func redactSecrets(cmd string) string {
	if cmd == "" {
		return cmd
	}

	cmd = privateKeyRegex.ReplaceAllString(cmd, "[REDACTED_PRIVATE_KEY]")
	cmd = awsAccessKeyRegex.ReplaceAllString(cmd, "[REDACTED_AWS_KEY]")
	cmd = bearerTokenRegex.ReplaceAllString(cmd, "Bearer [REDACTED_TOKEN]")

	cmd = genericTokenRegex.ReplaceAllStringFunc(cmd, func(match string) string {
		sub := genericTokenRegex.FindStringSubmatch(match)
		if len(sub) >= 3 {
			key := sub[1]
			return key + "=[REDACTED]"
		}
		return match
	})

	cmd = flagPasswordRegex.ReplaceAllStringFunc(cmd, func(match string) string {
		sub := flagPasswordRegex.FindStringSubmatch(match)
		if len(sub) >= 3 {
			flag := sub[1]
			return flag + " [REDACTED]"
		}
		return match
	})

	return cmd
}
