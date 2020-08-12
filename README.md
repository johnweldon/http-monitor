# http-monitor

Tool to validate expectations about endpoints.

Given a list of urls with associated methods, headers, and expected
results, this tool will periodically test all those expectations and
report anomalies.

Uses `go generate` because I wanted to figure out how to use that.
