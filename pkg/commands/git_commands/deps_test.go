package git_commands

import (
	"os"

	"github.com/go-errors/errors"
	gogit "github.com/jesseduffield/go-git/v5"
	"github.com/jesseduffield/lazygit/pkg/commands/git_config"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/spf13/afero"
)

type commonDeps struct {
	runner     *oscommands.FakeCmdObjRunner
	userConfig *config.UserConfig
	appState   *config.AppState
	gitVersion *GitVersion
	gitConfig  *git_config.FakeGitConfig
	getenv     func(string) string
	removeFile func(string) error
	common     *common.Common
	cmd        *oscommands.CmdObjBuilder
	fs         afero.Fs
	repoPaths  *RepoPaths
}

func buildGitCommon(deps commonDeps) *GitCommon {
	gitCommon := &GitCommon{}

	gitCommon.Common = deps.common
	if gitCommon.Common == nil {
		gitCommon.Common = common.NewDummyCommonWithUserConfigAndAppState(deps.userConfig, deps.appState)
	}

	if deps.fs != nil {
		gitCommon.Fs = deps.fs
	}

	if deps.repoPaths != nil {
		gitCommon.repoPaths = deps.repoPaths
	} else {
		gitCommon.repoPaths = MockRepoPaths(".git")
	}

	runner := deps.runner
	if runner == nil {
		runner = oscommands.NewFakeRunner(nil)
	}

	cmd := deps.cmd
	// gotta check deps.cmd because it's not an interface type and an interface value of nil is not considered to be nil
	if cmd == nil {
		cmd = oscommands.NewDummyCmdObjBuilder(runner)
	}
	gitCommon.cmd = cmd

	gitCommon.Common.SetUserConfig(deps.userConfig)
	if gitCommon.Common.UserConfig() == nil {
		gitCommon.Common.SetUserConfig(config.GetDefaultConfig())
	}

	gitCommon.version = deps.gitVersion
	if gitCommon.version == nil {
		gitCommon.version = &GitVersion{2, 0, 0, ""}
	}

	gitConfig := deps.gitConfig
	if gitConfig == nil {
		gitConfig = git_config.NewFakeGitConfig(nil)
	}

	gitCommon.repo = buildRepo()
	gitCommon.config = NewConfigCommands(gitCommon.Common, gitConfig, gitCommon.repo)

	getenv := deps.getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}

	removeFile := deps.removeFile
	if removeFile == nil {
		removeFile = func(string) error { return errors.New("unexpected call to removeFile") }
	}

	gitCommon.os = oscommands.NewDummyOSCommandWithDeps(oscommands.OSCommandDeps{
		Common:       gitCommon.Common,
		GetenvFn:     getenv,
		Cmd:          cmd,
		RemoveFileFn: removeFile,
		TempDir:      os.TempDir(),
	})

	return gitCommon
}

func buildRepo() *gogit.Repository {
	// TODO: think of a way to actually mock this out
	var repo *gogit.Repository
	return repo
}

func buildFileLoader(gitCommon *GitCommon) *FileLoader {
	return NewFileLoader(gitCommon, gitCommon.cmd, gitCommon.config)
}

func buildSubmoduleCommands(deps commonDeps) *SubmoduleCommands {
	gitCommon := buildGitCommon(deps)

	return NewSubmoduleCommands(gitCommon)
}

func buildCommitCommands(deps commonDeps) *CommitCommands {
	gitCommon := buildGitCommon(deps)
	return NewCommitCommands(gitCommon)
}

func buildWorkingTreeCommands(deps commonDeps) *WorkingTreeCommands {
	gitCommon := buildGitCommon(deps)
	submoduleCommands := buildSubmoduleCommands(deps)
	fileLoader := buildFileLoader(gitCommon)

	return NewWorkingTreeCommands(gitCommon, submoduleCommands, fileLoader)
}

func buildStashCommands(deps commonDeps) *StashCommands {
	gitCommon := buildGitCommon(deps)
	fileLoader := buildFileLoader(gitCommon)
	workingTreeCommands := buildWorkingTreeCommands(deps)

	return NewStashCommands(gitCommon, fileLoader, workingTreeCommands)
}

func buildRebaseCommands(deps commonDeps) *RebaseCommands {
	gitCommon := buildGitCommon(deps)
	workingTreeCommands := buildWorkingTreeCommands(deps)
	commitCommands := buildCommitCommands(deps)

	return NewRebaseCommands(gitCommon, commitCommands, workingTreeCommands)
}

func buildSyncCommands(deps commonDeps) *SyncCommands {
	gitCommon := buildGitCommon(deps)

	return NewSyncCommands(gitCommon)
}

func buildFileCommands(deps commonDeps) *FileCommands {
	gitCommon := buildGitCommon(deps)

	return NewFileCommands(gitCommon)
}

func buildBranchCommands(deps commonDeps) *BranchCommands {
	gitCommon := buildGitCommon(deps)

	return NewBranchCommands(gitCommon)
}

func buildFlowCommands(deps commonDeps) *FlowCommands {
	gitCommon := buildGitCommon(deps)

	return NewFlowCommands(gitCommon)
}
