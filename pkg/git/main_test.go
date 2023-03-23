package git

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
)

var (
	RootFolder       Folder
	MissingFolder    Folder
	EmptyFolder      Folder
	UnexpectedFolder Folder
	InitiatedFolder  Folder
	// Cannot simulate a situation that generates a UnknownFOlder situation
	//UnknownFolder    Folder
)

func TestMain(m *testing.M) {
	logger, _ := zap.NewDevelopment()
	log = logger.Sugar()
	if err := setupTesting(); err != nil {
		log.Panicf("error while setting up test requirements: %e", err)
	}
	defer func() {
		if err := teardownTesting(); err != nil {
			log.Fatalf("Error on teardown: %e", err)
		}
	}()
	exitcode := m.Run()
	_ = log.Sync()
	os.Exit(exitcode)
}

func writeFile(fileName string, data string) error {
	d1 := []byte(data)
	return os.WriteFile(fileName, d1, 0644)
}

func createTag(repo Folder, tag string, file string, data string) error {
	if err := writeFile(filepath.Join(string(repo), file), data); err != nil {
		return err
	} else if err = repo.RunGitCommand([]string{"add", "."}); err != nil {
		return err
	} else if err = repo.RunGitCommand([]string{"commit", "-m", "initial"}); err != nil {
		return err
	} else if err = repo.RunGitCommand([]string{"tag", tag}); err != nil {
		return err
	}
	return nil
}

func initRepo(repo Folder) error {
	if err := repo.RunGitCommand([]string{"init"}); err != nil {
		return err
	}

	for i := 1; i < 5; i++ {
		if err := createTag(repo, fmt.Sprintf("tag%d", i), fmt.Sprintf("foo%d", i), "bar"); err != nil {
			return err
		}
	}
	return nil
}

func setupTesting() error {
	var err error
	if fld, err := os.MkdirTemp("", "go_test_pgquartz"); err != nil {
		return fmt.Errorf("failed to create rootfolder")
	} else {
		RootFolder = Folder(fld)
	}
	MissingFolder = Folder(filepath.Join(RootFolder.String(), "missing"))
	if EmptyFolder, err = RootFolder.SubFolder("empty"); err != nil {
		return err
	} else if UnexpectedFolder, err = RootFolder.SubFolder("unexpected"); err != nil {
		return err
	} else if InitiatedFolder, err = RootFolder.SubFolder("initiated"); err != nil {
		return err
		// Missing folder and Empty folder require no further initialization
		// Unknown situations cannot be enforced, so skipping that entirely
	} else if err = writeFile(filepath.Join(UnexpectedFolder.String(), "dummy"), "bogus"); err != nil {
		return err
	} else if err = initRepo(InitiatedFolder); err != nil {
		return err
	}
	return nil
}

func teardownTesting() error {
	return os.RemoveAll(RootFolder.String())
}
