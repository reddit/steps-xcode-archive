package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-xcode/xcodebuild"
	cache "github.com/bitrise-io/go-xcode/xcodecache"
	"github.com/bitrise-io/go-xcode/xcpretty"
)

func runArchiveCommandWithRetry(archiveCmd *xcodebuild.CommandBuilder, useXcpretty bool, swiftPackagesPath string) (string, error) {
	output, err := runArchiveCommand(archiveCmd, useXcpretty)
	if err != nil && swiftPackagesPath != "" && strings.Contains(output, cache.SwiftPackagesStateInvalid) {
		log.Warnf("Archive failed, swift packages cache is in an invalid state, error: %s", err)
		log.RWarnf("xcode-archive", "swift-packages-cache-invalid", nil, "swift packages cache is in an invalid state")
		if err := os.RemoveAll(swiftPackagesPath); err != nil {
			return output, fmt.Errorf("failed to remove invalid Swift package caches, error: %s", err)
		}
		return runArchiveCommand(archiveCmd, useXcpretty)
	}
	return output, err
}

func runArchiveCommand(archiveCmd *xcodebuild.CommandBuilder, useXcpretty bool) (string, error) {
	if useXcpretty {
		xcprettyCmd := xcpretty.New(archiveCmd)

		logWithTimestamp(colorstring.Green, "$ %s", xcprettyCmd.PrintableCmd())
		fmt.Println()

		return xcprettyCmd.Run()
	}
	// Using xcodebuild
	logWithTimestamp(colorstring.Green, "$ %s", archiveCmd.PrintableCmd())
	fmt.Println()

	var output bytes.Buffer
	archiveRootCmd := archiveCmd.Command()
	archiveRootCmd.SetStdout(&output)
	archiveRootCmd.SetStderr(&output)

	var err error
	progress.SimpleProgress(".", time.Minute, func() {
		err = archiveRootCmd.Run()
	})

	return output.String(), err
}
