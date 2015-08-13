package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/juju/errors"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Trace(err)
}

func getPackageRootDir(repo string) (string, error) { // move backwards through the package name, looking for a .git/.hg dir to find the package "root"
	gopath := os.Getenv("GOPATH")
	resultPath := path.Join(gopath, "src", repo)

	parts := strings.Split(repo, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		repoPortion := path.Join(parts[:i]...)
		candidatePath := path.Join(gopath, "src", repoPortion)

		gitDir := path.Join(candidatePath, ".git")
		hgDir := path.Join(candidatePath, ".hg")
		bzrDir := path.Join(candidatePath, ".bzr")

		if exists, _ := pathExists(gitDir); exists {
			resultPath = candidatePath
			break
		}

		if exists, _ := pathExists(hgDir); exists {
			resultPath = candidatePath
			break
		}

		if exists, _ := pathExists(bzrDir); exists {
			resultPath = candidatePath
			break
		}
	}

	return resultPath, nil
}

func fetchPackage(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	gopath := os.Getenv("GOPATH")
	packageDir := path.Join(gopath, "src", getRealRepoPath(repo))

	if _, err := os.Stat(packageDir); err != nil {
		if os.IsNotExist(err) {
			var s *spinner.Spinner
			if Verbose {
				s = spinner.New(spinner.CharSets[SpinnerCharSet], SpinnerInterval)
				s.Prefix = fmt.Sprintf("fetching %s ", repo)
				s.Color("green")
				s.Start()
			}

			goGetCommand := []string{"go", "get", "-d", repo}
			goGetCmd := exec.Command(goGetCommand[0], goGetCommand[1:]...)
			err := goGetCmd.Run()

			if Verbose {
				s.Stop()
			}

			if err != nil {
				return errors.Annotatef(err, "failed cloning repo for package %s", repo)
			} else {
				if Verbose {
					fmt.Printf("\rfetching %s ... %s\n", repo, color.GreenString("done"))
				}
			}

			return nil
		}
	}

	defer func() {
		_ = os.Chdir(wd)
	}()

	packageDir, err = getPackageRootDir(getRealRepoPath(repo))
	if err != nil {
		return errors.Trace(err)
	}

	err = os.Chdir(packageDir)
	if err != nil {
		return errors.Trace(err)
	}

	var s *spinner.Spinner
	if Verbose {
		s = spinner.New(spinner.CharSets[SpinnerCharSet], SpinnerInterval)
		s.Prefix = fmt.Sprintf("refreshing %s ", repo)
		s.Color("green")
		s.Start()
	}

	var refreshCommand []string

	if exists, _ := pathExists(path.Join(packageDir, ".git")); exists {
		refreshCommand = []string{"git", "fetch", "--all"}
	} else if exists, _ := pathExists(path.Join(packageDir, ".hg")); exists {
		refreshCommand = []string{"hg", "pull"}
	} else if exists, _ := pathExists(path.Join(packageDir, ".bzr")); exists {
		refreshCommand = []string{"bzr", "pull"}
	} else if exists, _ := pathExists(path.Join(packageDir, ".svn")); exists {
		refreshCommand = []string{"svn", "up"}
	}

	if len(refreshCommand) > 0 {
		refreshOutput, err := exec.Command(refreshCommand[0], refreshCommand[1:]...).CombinedOutput()

		if Verbose {
			s.Stop()
			fmt.Printf("\rrefreshing %s ... %s\n", repo, color.GreenString("done"))
		}

		if err != nil {
			return errors.Annotatef(err, "failed updating repo for package %s, output: %s", repo, refreshOutput)
		}
	} else {
		if Verbose {
			s.Stop()
			fmt.Printf("\rrefreshing %s ... %s\n", repo, color.YellowString("skipped"))
		}
	}

	return nil
}

func fetchPackageDependencies(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	gopath := os.Getenv("GOPATH")
	packageDir := path.Join(gopath, "src", getRealRepoPath(repo))

	defer func() {
		_ = os.Chdir(wd)
	}()

	err = os.Chdir(packageDir)
	if err != nil {
		return errors.Trace(err)
	}

	var s *spinner.Spinner

	if Verbose {
		s = spinner.New(spinner.CharSets[SpinnerCharSet], SpinnerInterval)
		s.Prefix = fmt.Sprintf("  - fetching dependencies for %s ", repo)
		s.Color("green")
		s.Start()
	}

	goGetCommand := []string{"go", "get", "-u", "-d", "./..."}
	goGetOutput, err := exec.Command(goGetCommand[0], goGetCommand[1:]...).CombinedOutput()

	if Verbose {
		s.Stop()
		fmt.Printf("\r  - fetching dependencies for %s ... %s\n", repo, color.GreenString("done"))
	}

	if err != nil {
		return errors.Annotatef(err, "failed fetching dependencies for package %s, output: %s", repo, goGetOutput)
	}

	return nil
}

func buildPackage(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	gopath := os.Getenv("GOPATH")
	packageDir := path.Join(gopath, "src", getRealRepoPath(repo))

	defer func() {
		_ = os.Chdir(wd)
	}()

	err = os.Chdir(packageDir)
	if err != nil {
		return errors.Trace(err)
	}

	var s *spinner.Spinner

	if Verbose {
		s = spinner.New(spinner.CharSets[SpinnerCharSet], SpinnerInterval)
		s.Prefix = fmt.Sprintf("  - building package %s ", repo)
		s.Color("green")
		s.Start()
	}

	goBuildCommand := []string{"go", "build", repo}
	goBuildOutput, err := exec.Command(goBuildCommand[0], goBuildCommand[1:]...).CombinedOutput()

	if Verbose {
		s.Stop()
		fmt.Printf("\r  - building package %s ... %s\n", repo, color.GreenString("done"))
	}

	if err != nil {
		return errors.Annotatef(err, "failed building package %s, error: %s, output: %s", repo, goBuildOutput)
	}

	return nil
}

func installPackage(repo string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	gopath := os.Getenv("GOPATH")
	packageDir := path.Join(gopath, "src", getRealRepoPath(repo))

	defer func() {
		_ = os.Chdir(wd)
	}()

	err = os.Chdir(packageDir)
	if err != nil {
		return errors.Trace(err)
	}

	var s *spinner.Spinner

	if Verbose {
		s = spinner.New(spinner.CharSets[SpinnerCharSet], SpinnerInterval)
		s.Prefix = fmt.Sprintf("  - installing package %s ", repo)
		s.Color("green")
		s.Start()
	}

	goInstallCommand := []string{"go", "install", repo}
	goInstallOutput, err := exec.Command(goInstallCommand[0], goInstallCommand[1:]...).CombinedOutput()

	if Verbose {
		s.Stop()
		fmt.Printf("\r  - installing package %s ... %s\n", repo, color.GreenString("done"))
	}

	if err != nil {
		return errors.Annotatef(err, "failed installing package %s, error: %s, output: %s", repo, goInstallOutput)
	}

	return nil
}

func setPackageVersion(repo string, version string, humanVersion string) error {
	if version == "" {
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	packageDir, err := getPackageRootDir(getRealRepoPath(repo))
	if err != nil {
		return errors.Trace(err)
	}

	defer func() {
		_ = os.Chdir(wd)
	}()

	err = os.Chdir(packageDir)
	if err != nil {
		return errors.Trace(err)
	}

	var checkoutCommand []string
	if exists, _ := pathExists(".git"); exists {
		checkoutCommand = []string{"git", "checkout", version}
	} else if exists, _ := pathExists(".hg"); exists {
		checkoutCommand = []string{"hg", "update", "-c", version}
	} else if exists, _ := pathExists(".bzr"); exists {
		if version != "" {
			checkoutCommand = []string{"bzr", "update", "-r", version}
		} else {
			checkoutCommand = []string{"bzr", "update"}
		}
	} else {
		if Verbose {
			fmt.Printf("  - setting version of %s to %s (resolved as %s) ... %s\n", repo, humanVersion, version, color.GreenString("skipped, unknown repo type"))
		}
		return nil
	}

	var s *spinner.Spinner

	if Verbose {
		s = spinner.New(spinner.CharSets[SpinnerCharSet], SpinnerInterval)
		s.Prefix = fmt.Sprintf("  - setting version of %s to %s (resolved as %s) ", repo, humanVersion, version)
		s.Color("green")
		s.Start()
	}

	checkoutOutput, err := exec.Command(checkoutCommand[0], checkoutCommand[1:]...).CombinedOutput()

	if Verbose {
		s.Stop()
		fmt.Printf("\r  - setting version of %s to %s (resolved as %s) ... %s\n", repo, humanVersion, version, color.GreenString("done"))
	}

	if err != nil {
		return errors.Annotatef(err, "failed setting version of package %s, error: %s, output: %s", repo, checkoutOutput)
	}

	return nil
}

func countNonEmptyStrings(ar []string) int {
	counter := 0

	for _, el := range ar {
		if el != "" {
			counter += 1
		}
	}

	return counter
}

func getRealRepoPath(pack string) string {
	if strings.HasSuffix(pack, "...") {
		return strings.Replace(pack, "/...", "", -1)
	} else {
		return pack
	}
}

type PackageRecencyInfo struct {
	LatestCommit         string
	LatestUpstreamCommit string
	InstalledCommit      string
	UpstreamDiffCount    int
	InstalledDiffCount   int
}

func checkPackageRecency(pack Package) (bool, PackageRecencyInfo, error) { // bool = needsUpdate
	NilInfo := PackageRecencyInfo{}

	repo := getRealRepoPath(pack.Repo)
	version, err := getLatestVersionMatchingPattern(pack.Repo, pack.Version)

	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	gopath := os.Getenv("GOPATH")
	packageDir := path.Join(gopath, "src", repo)

	if exists, _ := pathExists(packageDir); !exists {
		return true, NilInfo, nil
	}

	packageDir, err = getPackageRootDir(repo)
	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	defer func() {
		_ = os.Chdir(wd)
	}()

	err = os.Chdir(packageDir)
	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	var repoType string

	if exists, _ := pathExists(".git"); exists {
		repoType = "git"
	} else if exists, _ := pathExists(".hg"); exists {
		repoType = "hg"
	} else {
		return true, NilInfo, nil // force an update
	}

	var getVersionCommand, getHEADCommand, getUpstreamVersionCommand, getUpstreamDiffCommand, getInstalledDiffCommand []string

	if repoType == "git" {
		getVersionCommand = []string{"git", "rev-parse", "-q", "--verify", version}
		getHEADCommand = []string{"git", "rev-parse", "-q", "--verify", "HEAD"}
		getUpstreamVersionCommand = []string{"git", "rev-parse", "-q", "--verify", "origin/master"}
		getUpstreamDiffCommand = []string{"git", "log", "HEAD..origin/master", "--pretty=oneline"}
		getInstalledDiffCommand = []string{"git", "log", fmt.Sprintf("HEAD..%s", version), "--pretty=oneline"}
	} else if repoType == "hg" {
		getVersionCommand = []string{"hg", "identify", "-ir", version}
		getHEADCommand = []string{"hg", "identify", "-i"}
		getUpstreamVersionCommand = []string{"hg", "identify", "-ir", "tip"} // imperfect, but there's no git equivalent to this
		getUpstreamDiffCommand = []string{"echo"}
		getInstalledDiffCommand = []string{"echo"} // can't really even approximate this
	}

	getVersionOutput, err := exec.Command(getVersionCommand[0], getVersionCommand[1:]...).Output()
	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	getUpstreamVersionOutput, err := exec.Command(getUpstreamVersionCommand[0], getUpstreamVersionCommand[1:]...).Output()
	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	getHEADOutput, err := exec.Command(getHEADCommand[0], getHEADCommand[1:]...).Output()
	if err != nil {
		return false, NilInfo, errors.Trace(err)
	}

	upstreamDiffCount := 0
	getUpstreamDiffOutput, err := exec.Command(getUpstreamDiffCommand[0], getUpstreamDiffCommand[1:]...).CombinedOutput()
	if err == nil {
		upstreamDiffCount = countNonEmptyStrings(strings.Split(strings.TrimSpace(string(getUpstreamDiffOutput)), "\n"))
	}

	installedDiffCount := 0
	getInstalledDiffOutput, err := exec.Command(getInstalledDiffCommand[0], getInstalledDiffCommand[1:]...).CombinedOutput()
	if err == nil {
		installedDiffCount = countNonEmptyStrings(strings.Split(strings.TrimSpace(string(getInstalledDiffOutput)), "\n"))
	}

	versionString := strings.TrimSpace(string(getVersionOutput))
	HEADString := strings.TrimSpace(string(getHEADOutput))
	upstreamVersionString := strings.TrimSpace(string(getUpstreamVersionOutput))

	recencyInfo := PackageRecencyInfo{
		LatestCommit:         versionString,
		LatestUpstreamCommit: upstreamVersionString,
		InstalledCommit:      HEADString,
		UpstreamDiffCount:    upstreamDiffCount,
		InstalledDiffCount:   installedDiffCount,
	}

	pkgPath := fmt.Sprintf("%s.a", path.Join(gopath, "pkg", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH), repo))

	if exists, _ := pathExists(pkgPath); !exists {
		return true, recencyInfo, nil
	}

	if versionString != HEADString {
		if pack.LockedVersion != HEADString {
			if version == "" {
				return false, recencyInfo, nil
			} else {
				return true, recencyInfo, nil
			}
		} else {
			return false, recencyInfo, nil
		}
	} else {
		if pack.LockedVersion != "" && pack.LockedVersion != HEADString {
			return true, recencyInfo, nil
		} else {
			return false, recencyInfo, nil
		}
	}

	return false, NilInfo, nil
}

func parsePackage(packString string) Package {
	parts := strings.Split(packString, "@")
	pack := Package{}

	if len(parts) == 2 {
		pack.Repo = parts[0]
		pack.Version = parts[1]
	} else {
		pack.Repo = parts[0]
	}

	repoParts := strings.Split(pack.Repo, "/")
	if len(repoParts) == 2 {
		if !strings.Contains(repoParts[0], ".") {
			// github shorthand
			pack.Repo = fmt.Sprintf("github.com/%s", pack.Repo)
		}
	}

	return pack
}

func installPackagesFromBunchfile(b *BunchFile, forceUpdate bool, checkUpstream bool, respectLocked bool) error {
	return installPackages(b.Packages, false, forceUpdate, checkUpstream, respectLocked)
}

func installPackagesFromRepoStrings(packageStrings []string, installGlobally bool, forceUpdate bool, checkUpstream bool, respectLocked bool) error {
	packages := make([]Package, len(packageStrings))
	for i, packString := range packageStrings {
		packages[i] = parsePackage(packString)
	}

	return installPackages(packages, installGlobally, forceUpdate, checkUpstream, respectLocked)
}

func installPackages(packages []Package, installGlobally bool, forceUpdate bool, checkUpstream bool, respectLocked bool) error {
	if !installGlobally {
		err := setVendorEnv()
		if err != nil {
			return errors.Trace(err)
		}
	}

	gopath := os.Getenv("GOPATH")

	anyNeededUpdate := false
	packageNeedsUpdate := make(map[string]bool)

	for _, pack := range packages {
		if pack.IsLink {
			repoPath := path.Join(gopath, "src", pack.Repo)

			if exists, _ := pathExists(repoPath); !exists {
				err := os.MkdirAll(filepath.Dir(repoPath), 0755)
				if err != nil {
					return errors.Trace(err)
				}

				err = os.Symlink(pack.LinkTarget, path.Join(gopath, "src", pack.Repo))
				if err != nil {
					return errors.Trace(err)
				}

				if !pack.IsSelf {
					fmt.Printf("\rsetting up local package %s ... %s      \n", pack.Repo, color.GreenString("done"))
				} else {
					fmt.Printf("\rsetting up %s link ... %s      \n", pack.Repo, color.GreenString("done"))
				}
			}

			continue
		}

		needsUpdate, _, err := checkPackageRecency(pack)
		if err != nil {
			return errors.Trace(err)
		}

		if needsUpdate {
			packageNeedsUpdate[pack.Repo] = true
			anyNeededUpdate = true
		}

		if (needsUpdate || forceUpdate) && checkUpstream {
			if !Verbose {
				fmt.Printf("fetching %s ... ", pack.Repo)
			}

			err = fetchPackage(pack.Repo)
			if err != nil {
				return errors.Trace(err)
			}

			err = fetchPackageDependencies(pack.Repo)
			if err != nil {
				return errors.Trace(err)
			}

			if Verbose {
				fmt.Println("")
			} else {
				fmt.Printf("\rfetching %s ... %s      \n", pack.Repo, color.GreenString("done"))
			}
		}
	}

	for _, pack := range packages {
		needsUpdate := packageNeedsUpdate[pack.Repo]

		if needsUpdate || forceUpdate {
			if Verbose {
				fmt.Printf("installing %s ... \n", pack.Repo)
			} else {
				fmt.Printf("installing %s ... ", pack.Repo)
			}

			version := pack.Version

			if !pack.IsLink {
				var err error
				version, err = getLatestVersionMatchingPattern(pack.Repo, pack.Version)
				if err != nil {
					return errors.Trace(err)
				}
			}

			if pack.LockedVersion != "" && respectLocked {
				version = pack.LockedVersion
			}

			if !pack.IsLink {
				err := setPackageVersion(pack.Repo, version, pack.Version)
				if err != nil {
					return errors.Trace(err)
				}
			}

			if !pack.IsSelf {
				err := buildPackage(pack.Repo)
				if err != nil {
					return errors.Trace(err)
				}

				err = installPackage(pack.Repo)
				if err != nil {
					return errors.Trace(err)
				}
			}

			if Verbose {
				fmt.Print(color.GreenString("\rsuccessfully installed %s                 \n\n", pack.Repo))
			} else {
				fmt.Printf("\rinstalling %s ... %s      \n", pack.Repo, color.GreenString("done"))
			}

		} else {
			if Verbose {
				fmt.Print(color.YellowString("skipping %s, up to date                 \n", pack.Repo))
			}
		}
	}

	if !anyNeededUpdate && !Verbose && !forceUpdate {
		color.Green("up to date (use 'bunch update' to force update)")
	}

	return nil
}

type GoList struct {
	Name        string
	Doc         string
	ImportPath  string
	Imports     []string
	TestImports []string
	Deps        []string
}

func isEmptyDir(name string) (bool, error) {
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	return len(entries) == 0, nil
}

func cleanEmpties(emptyPath string) error {
	higherPath := filepath.Dir(emptyPath)
	if empty, _ := isEmptyDir(higherPath); empty {
		err := os.Remove(higherPath)
		if err != nil {
			return errors.Trace(err)
		}

		highestPath := filepath.Dir(higherPath)
		if empty, _ := isEmptyDir(highestPath); empty {
			err := os.Remove(highestPath)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}

func removePackage(pack string) error {
	gopath := os.Getenv("GOPATH")

	srcPath := path.Join(gopath, "src", pack)
	if exists, _ := pathExists(srcPath); exists {
		err := os.RemoveAll(srcPath)
		if err != nil {
			return errors.Trace(err)
		}
	}

	err := cleanEmpties(srcPath)
	if err != nil {
		return errors.Trace(err)
	}

	archPath := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	pkgPath := fmt.Sprintf("%s.a", path.Join(gopath, "pkg", archPath, pack))
	if exists, _ := pathExists(pkgPath); exists {
		err := os.Remove(pkgPath)
		if err != nil {
			return errors.Trace(err)
		}
	}

	err = cleanEmpties(pkgPath)
	if err != nil {
		return errors.Trace(err)
	}

	_, binFile := path.Split(pack)

	if binFile != "" {
		binPath := path.Join(gopath, "bin", binFile)
		if exists, _ := pathExists(binPath); exists {
			err := os.Remove(binPath)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}

func removePackages(packages []string, bunch *BunchFile, removeGlobally bool) error {
	if !removeGlobally {
		err := setVendorEnv()
		if err != nil {
			return errors.Trace(err)
		}
	}

	gopath := os.Getenv("GOPATH")

	allPackages := make(map[string]bool)
	packagesUsed := make(map[string][]string)

	combinedPackagesList := packages
	for _, packData := range bunch.Packages {
		combinedPackagesList = append(combinedPackagesList, packData.Repo)
	}

	for _, pack := range combinedPackagesList {
		if exists, _ := pathExists(path.Join(gopath, "src", pack)); !exists {
			continue
		}

		goListCommand := []string{"go", "list", "--json", pack}
		output, err := exec.Command(goListCommand[0], goListCommand[1:]...).Output()
		if err != nil {
			return errors.Trace(err)
		}

		packageInfo := GoList{}
		err = json.Unmarshal(output, &packageInfo)

		if err != nil {
			return errors.Trace(err)
		}

		removingPackage := false
		for _, removePack := range packages {
			if removePack == pack {
				removingPackage = true
			}
		}

		if !removingPackage {
			packagesUsed[pack] = append(packagesUsed[pack], "app")
		}

		for _, dep := range packageInfo.Deps {
			srcPath := path.Join(gopath, "src", dep)
			if exists, _ := pathExists(srcPath); exists {
				allPackages[dep] = true

				if !removingPackage {
					packagesUsed[dep] = append(packagesUsed[dep], pack)
				}
			}
		}

		allPackages[pack] = true
	}

	for _, pack := range packages {
		if len(packagesUsed[pack]) > 0 {
			color.Red("unable to remove package %s, is depended on by %s", pack, strings.Join(packagesUsed[pack], ", "))
		}
	}

	for pack, _ := range allPackages {
		if len(packagesUsed[pack]) == 0 {
			fmt.Printf("removing package %s ...", pack)
			err := removePackage(pack)

			if err != nil {
				return errors.Trace(err)
			}

			fmt.Printf("\rremoving package %s ... %s      \n", pack, color.GreenString("done"))
		}
	}

	return nil
}

func isRootPackageUsed(packagesUsed map[string]bool, packName string) bool {
	for pack, _ := range packagesUsed {
		if strings.HasPrefix(pack, packName) {
			return true
		}
	}

	return false
}

func prunePackages(bunch *BunchFile) error {
	err := setVendorEnv()
	if err != nil {
		return errors.Trace(err)
	}

	gopath := os.Getenv("GOPATH")

	packagesUsed := make(map[string]bool)

	for _, packInfo := range bunch.Packages {
		pack := packInfo.Repo

		if exists, _ := pathExists(path.Join(gopath, "src", pack)); !exists {
			continue
		}

		goListCommand := []string{"go", "list", "--json", pack}
		output, err := exec.Command(goListCommand[0], goListCommand[1:]...).Output()
		if err != nil {
			return errors.Trace(err)
		}

		packageInfo := GoList{}
		err = json.Unmarshal(output, &packageInfo)

		if err != nil {
			return errors.Trace(err)
		}

		packagesUsed[pack] = true

		for _, dep := range packageInfo.Deps {
			srcPath := path.Join(gopath, "src", dep)
			if exists, _ := pathExists(srcPath); exists {
				packagesUsed[dep] = true
			}
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return errors.Trace(err)
	}

	err = os.Chdir(path.Join(gopath, "src"))
	if err != nil {
		return errors.Trace(err)
	}

	packFiles := []string{}
	err = filepath.Walk(".", func(packPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		gitExists, _ := pathExists(path.Join(packPath, ".git"))
		hgExists, _ := pathExists(path.Join(packPath, ".hg"))
		bzrExists, _ := pathExists(path.Join(packPath, ".bzr"))

		if gitExists || hgExists || bzrExists {
			packFiles = append(packFiles, packPath)
		}

		return nil
	})
	if err != nil {
		return errors.Trace(err)
	}

	err = os.Chdir(wd)
	if err != nil {
		return errors.Trace(err)
	}

	for _, pack := range packFiles {
		if !packagesUsed[pack] && !isRootPackageUsed(packagesUsed, pack) {
			fmt.Printf("removing package %s ...", pack)
			err := removePackage(pack)

			if err != nil {
				return errors.Trace(err)
			}

			fmt.Printf("\rremoving package %s ... %s      \n", pack, color.GreenString("done"))
		}
	}

	return nil
}

func gitShort(fullhash string) string {
	if len(fullhash) < 8 {
		return fullhash
	} else {
		return fullhash[:7]
	}
}

func commitsPlural(n int) string {
	if n == 1 {
		return "1 commit"
	} else {
		return fmt.Sprintf("%d commits", n)
	}
}

func checkOutdatedPackages(b *BunchFile) error {
	err := setVendorEnv()
	if err != nil {
		return errors.Trace(err)
	}

	for _, pack := range b.Packages {
		if pack.IsSelf {
			continue
		}

		fmt.Printf("package %s ... ", pack.Repo)

		err := fetchPackage(pack.Repo)
		if err != nil {
			return errors.Trace(err)
		}

		needsUpdate, recency, err := checkPackageRecency(pack)
		if err != nil {
			return errors.Trace(err)
		}

		if !needsUpdate {
			if recency.UpstreamDiffCount == 0 {
				fmt.Printf("\rpackage %s ... %s\n", pack.Repo, color.GreenString("up to date"))
			} else {
				if pack.LockedVersion == "" {
					fmt.Printf("\rpackage %s ... %s by %s, current is %6s, latest is %6s\n", pack.Repo, color.YellowString("behind upstream"), commitsPlural(recency.UpstreamDiffCount), gitShort(recency.InstalledCommit), gitShort(recency.LatestUpstreamCommit))
				} else {
					fmt.Printf("\rpackage %s ... %s by %s, current is %6s, latest is %6s\n", pack.Repo, color.YellowString("locked, but behind upstream"), commitsPlural(recency.UpstreamDiffCount), gitShort(recency.InstalledCommit), gitShort(recency.LatestUpstreamCommit))
				}
			}
		} else {
			if recency.InstalledDiffCount == 0 {
				fmt.Printf("\rpackage %s ... %s\n", pack.Repo, color.GreenString("up to date"))
			} else {
				if pack.LockedVersion == "" {
					fmt.Printf("\rpackage %s ... %s by %s, current is %6s, latest is %6s\n", pack.Repo, color.RedString("outdated"), commitsPlural(recency.InstalledDiffCount), gitShort(recency.InstalledCommit), gitShort(recency.LatestCommit))
				} else {
					fmt.Printf("\rpackage %s ... %s by %s, current is %6s, latest is %6s\n", pack.Repo, color.YellowString("locked, but outdated"), commitsPlural(recency.InstalledDiffCount), gitShort(recency.InstalledCommit), gitShort(recency.LatestCommit))
				}
			}
		}
	}

	return nil
}

func lockPackages(b *BunchFile) error {
	err := setVendorEnv()
	if err != nil {
		return errors.Trace(err)
	}

	lockList := make(map[string]string)

	for _, pack := range b.Packages {
		if pack.IsLink {
			continue
		}

		_, recency, err := checkPackageRecency(pack)
		if err != nil {
			return errors.Trace(err)
		}

		lockList[pack.Repo] = recency.LatestCommit
	}

	jsonOut, err := json.MarshalIndent(lockList, "", "    ")
	if err != nil {
		return errors.Trace(err)
	} else {
		err = ioutil.WriteFile("Bunchfile.lock", append(jsonOut, '\n'), 0644)
		if err != nil {
			return errors.Trace(err)
		}

		color.Green("Bunchfile.lock generated successfully")
	}

	return nil
}
