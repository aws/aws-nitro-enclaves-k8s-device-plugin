// Copyright 2026 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestVersionFlag builds the binary with -ldflags -X injection and runs it
// with -version to verify the full build-time injection contract.
func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		buildDate string
		wantRegex string
	}{
		{
			name:      "injected release values",
			version:   "0.4.1",
			buildDate: "2026-04-22T17:33:56Z",
			wantRegex: `^k8s-ne-device-plugin version 0\.4\.1 \(built: 2026-04-22T17:33:56Z\)\n$`,
		},
		{
			name:      "defaults when no injection",
			version:   "",
			buildDate: "",
			wantRegex: `^k8s-ne-device-plugin version dev \(built: unknown\)\n$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			binPath := filepath.Join(t.TempDir(), "k8s-ne-device-plugin")

			args := []string{"build", "-o", binPath}
			if tc.version != "" || tc.buildDate != "" {
				ldflags := "-X main.version=" + tc.version + " -X main.buildDate=" + tc.buildDate
				args = append(args, "-ldflags", ldflags)
			}
			args = append(args, ".")

			buildCmd := exec.Command("go", args...)
			if out, err := buildCmd.CombinedOutput(); err != nil {
				t.Fatalf("go build failed: %v\n%s", err, out)
			}

			out, err := exec.Command(binPath, "-version").CombinedOutput()
			if err != nil {
				t.Fatalf("running binary with -version failed: %v\n%s", err, out)
			}

			got := string(out)
			if !regexp.MustCompile(tc.wantRegex).MatchString(got) {
				t.Errorf("unexpected output\n got: %q\nwant regex: %q", got, tc.wantRegex)
			}
			if strings.Contains(got, "Starting K8s Nitro Enclaves") {
				t.Errorf("-version output should not include startup log line, got: %q", got)
			}
		})
	}
}
