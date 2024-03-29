// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package release

import (
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd/core"
	buildconfigs "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/local/buildconfigs"
	"github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/logger"
)

type ReleaseOptions struct {
	ui          cmdcore.AuthoringUI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	pkgVersion     string
	chdir          string
	outputLocation string
	debug          bool
	tag            string
}

const (
	defaultArtifactDir = "carvel-artifacts"
)

func NewReleaseOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *ReleaseOptions {
	return &ReleaseOptions{ui: cmdcore.NewAuthoringUIImpl(ui), depsFactory: depsFactory, logger: logger}
}

func NewReleaseCmd(o *ReleaseOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release package",
		RunE:  func(cmd *cobra.Command, args []string) error { return o.Run() },
	}

	cmd.Flags().StringVarP(&o.pkgVersion, "version", "v", "", "Version to be released")
	cmd.Flags().StringVar(&o.chdir, "chdir", "", "Working directory with package-build and other config")
	cmd.Flags().StringVar(&o.outputLocation, "copy-to", defaultArtifactDir, "Output location for artifacts")
	cmd.Flags().BoolVar(&o.debug, "debug", false, "Print verbose debug output")
	cmd.Flags().StringVarP(&o.tag, "tag", "t", "", "Tag pushed with imgpkg bundle (default build-<TIMESTAMP>)")

	return cmd
}

func (o *ReleaseOptions) Run() error {
	if o.chdir != "" {
		err := os.Chdir(o.chdir)
		if err != nil {
			return err
		}
	}

	o.printPrerequisites()

	appBuild, err := buildconfigs.NewAppBuildFromFile(buildconfigs.AppBuildFileName)
	if err != nil {
		return err
	}

	builderOpts := AppSpecBuilderOpts{
		BuildTemplate: appBuild.GetAppSpec().Template,
		BuildDeploy:   appBuild.GetAppSpec().Deploy,
		BuildExport:   appBuild.GetExport(),
		Debug:         o.debug,
		BundleTag:     o.tag,
	}
	_, err = NewAppSpecBuilder(o.depsFactory, o.logger, o.ui, builderOpts).Build()
	if err != nil {
		return err
	}

	// TODO: Write app resources

	return nil
}

func (o *ReleaseOptions) printPrerequisites() {
	o.ui.PrintHeaderText("Pre-requisites")
	o.ui.PrintInformationalText("1. Host is authorized to push images to a registry (can be set up using `docker login`)\n" +
		"2. `app init` ran successfully.")
}
