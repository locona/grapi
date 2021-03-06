package usecase

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/izumin5210/grapi/pkg/grapicmd/internal/module"
	"github.com/izumin5210/grapi/pkg/grapicmd/protoc"
	"github.com/izumin5210/grapi/pkg/grapicmd/util/fs"
)

// ExecuteProtocUsecase is an useecase interface for executing protoc module.
type ExecuteProtocUsecase interface {
	Perform() error
	InstallPlugins() error
	ExecuteProtoc() error
}

type executeProtocUsecase struct {
	cfg             *protoc.Config
	fs              afero.Fs
	ui              module.UI
	commandFactory  module.CommandFactory
	rootDir, binDir string
}

// NewExecuteProtocUsecase returns an new ExecuteProtocUsecase implementation instance.
func NewExecuteProtocUsecase(cfg *protoc.Config, fs afero.Fs, ui module.UI, commandFactory module.CommandFactory, rootDir string) ExecuteProtocUsecase {
	return &executeProtocUsecase{
		cfg:            cfg,
		fs:             fs,
		ui:             ui,
		commandFactory: commandFactory,
		rootDir:        rootDir,
		binDir:         filepath.Join(rootDir, "bin"),
	}
}

func (u *executeProtocUsecase) Perform() error {
	u.ui.Section("Execute protoc")
	u.ui.Subsection("Install plugins")
	err := errors.WithStack(u.InstallPlugins())
	if err != nil {
		return err
	}
	u.ui.Subsection("Execute protoc")
	return errors.WithStack(u.ExecuteProtoc())
}

func (u *executeProtocUsecase) InstallPlugins() error {
	if err := fs.CreateDirIfNotExists(u.fs, u.binDir); err != nil {
		return errors.WithStack(err)
	}
	var errs []error
	for _, plugin := range u.cfg.Plugins {
		ok, err := u.installPlugin(plugin)
		if err != nil {
			errs = append(errs, err)
			u.ui.ItemFailure(plugin.BinName())
		} else if !ok {
			u.ui.ItemSkipped(plugin.BinName())
		} else {
			u.ui.ItemSuccess(plugin.BinName())
		}
	}
	if len(errs) > 0 {
		for _, err := range errs {
			u.ui.Error(err.Error())
		}
		return errors.New("failed to install protoc plugins")
	}
	return nil
}

func (u *executeProtocUsecase) ExecuteProtoc() error {
	protoFiles, err := u.cfg.ProtoFiles(u.fs, u.rootDir)
	if err != nil {
		return errors.WithStack(err)
	}
	var errs []error
	for _, path := range protoFiles {
		err = u.executeProtoc(path)
		relPath, _ := filepath.Rel(u.rootDir, path)
		if err == nil {
			u.ui.ItemSuccess(relPath)
		} else {
			errs = append(errs, err)
			u.ui.ItemFailure(relPath)
		}
	}
	if len(errs) > 0 {
		for _, err := range errs {
			u.ui.Error(err.Error())
		}
		return errors.New("failed to execute protoc")
	}
	return nil
}

func (u *executeProtocUsecase) installPlugin(plugin *protoc.Plugin) (bool, error) {
	binPath := filepath.Join(u.binDir, plugin.BinName())
	if ok, err := afero.Exists(u.fs, binPath); err != nil {
		return false, errors.Wrapf(err, "failed to get %q binary", plugin.BinName())
	} else if ok {
		return false, nil
	}
	dir := filepath.Join(u.rootDir, plugin.Path)
	if ok, _ := afero.DirExists(u.fs, dir); !ok {
		return false, errors.Errorf("%s is not found", plugin.Path)
	}
	cmd := u.commandFactory.Create([]string{"go", "install", "."})
	out, err := cmd.SetDir(dir).AddEnv("GOBIN", u.binDir).Exec()
	if err != nil {
		return false, errors.Wrapf(err, "failed to execute module: %s", string(out))
	}
	return true, nil
}

func (u *executeProtocUsecase) executeProtoc(protoPath string) error {
	outDir, err := u.cfg.OutDirOf(u.rootDir, protoPath)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = fs.CreateDirIfNotExists(u.fs, outDir); err != nil {
		return errors.WithStack(err)
	}
	cmds, err := u.cfg.Commands(u.rootDir, protoPath)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, cmd := range cmds {
		out, err := u.commandFactory.Create(cmd).AddEnv("PATH", u.binDir+string(filepath.ListSeparator)+os.Getenv("PATH")).SetDir(u.rootDir).Exec()
		if err != nil {
			return errors.Wrapf(err, "failed to execute module: %s", string(out))
		}
	}
	return nil
}
