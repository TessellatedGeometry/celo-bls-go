package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const WRITE_PERMS = 0777
const LIB_NAME = "libbls_snark_sys.a"

type Platform struct {
	Name string
	BuildDirective string
	LinkageDirective string
	LibDirectory string
}

func main() {
	sourceDir, _ := os.Args[1], os.Args[2]
	platformsDirPath := path.Join(sourceDir, "platforms")
	platformsPath := path.Join(platformsDirPath, "platforms.json")
	platformsContent, err := ioutil.ReadFile(platformsPath)
	panicIfError(err)
	var platforms map[string][]Platform
	err = json.Unmarshal(platformsContent, &platforms)
	panicIfError(err)

	for _, module := range []string{"bls", "snark"} {
		moduleGoOriginalPath := path.Join(platformsDirPath, fmt.Sprintf("%s.go.template", module))
		moduleHOriginalPath := path.Join(platformsDirPath, fmt.Sprintf("%s.h", module))
		libsDirOriginalPath := path.Join(sourceDir, "libs")

		reposDirPath := path.Join(platformsDirPath, "repos")
		goModTemplate, err := ioutil.ReadFile(path.Join(platformsDirPath, "go.mod.template"))
		panicIfError(err)
		modulePlatformTemplate, err := ioutil.ReadFile(path.Join(platformsDirPath, "platform.template"))
		panicIfError(err)
		for pkg, platformsForPkg := range platforms {
			var pkgBuildDirectives []string
			repoPath := path.Join(reposDirPath, fmt.Sprintf("celo-bls-go-%s", pkg))
			goMod := strings.Replace(string(goModTemplate), "{PACKAGE}", pkg, 1)
			goModPath := path.Join(repoPath, "go.mod")
			err = writeFile(goModPath, []byte(goMod))
			panicIfError(err)
			moduleDirPath := path.Join(repoPath, module)
			moduleGoPath := path.Join(moduleDirPath, fmt.Sprintf("%s.go", module))
			err = copyFile(moduleGoOriginalPath, moduleGoPath)
			panicIfError(err)
			moduleHPath := path.Join(moduleDirPath, fmt.Sprintf("%s.h", module))
			err = copyFile(moduleHOriginalPath, moduleHPath)
			panicIfError(err)
			for _, platform := range platformsForPkg {
				modulePlatformPath := path.Join(moduleDirPath, fmt.Sprintf("%s_%s.go", module, platform.Name))
				modulePlatform := strings.Replace(string(modulePlatformTemplate), "{BUILD_DIRECTIVE}", platform.BuildDirective, 1)
				modulePlatform = strings.Replace(modulePlatform, "{LINKAGE_DIRECTIVE}", platform.LinkageDirective, 1)
				modulePlatform = strings.Replace(modulePlatform, "{MODULE}", module, 1)
				err = writeFile(modulePlatformPath, []byte(modulePlatform))
				panicIfError(err)
				pkgBuildDirectives = append(pkgBuildDirectives, platform.BuildDirective)

				if module == "bls" {
					libsDirPath := path.Join(repoPath, "libs")
					libOriginalPath := path.Join(libsDirOriginalPath, platform.LibDirectory, LIB_NAME)
					libPath := path.Join(libsDirPath, platform.LibDirectory, LIB_NAME)
					err = copyFile(libOriginalPath, libPath)
					panicIfError(err)
				}
			}
			pkgRouterTemplatePath := path.Join(platformsDirPath, fmt.Sprintf("%s_router.go.template", module))
			pkgRouterTemplate, err := ioutil.ReadFile(pkgRouterTemplatePath)
			panicIfError(err)
			pkgRouter := strings.Replace(string(pkgRouterTemplate), "{PACKAGE}", pkg, 1)
			pkgRouter = strings.Replace(pkgRouter, "{BUILD_DIRECTIVE}", strings.Join(pkgBuildDirectives, " "), 1)

			modulePath := path.Join(sourceDir, module)
			pkgRouterPath := path.Join(modulePath, fmt.Sprintf("%s_%s.go", module, pkg))
			err = writeFile(pkgRouterPath, []byte(pkgRouter))
		}
	}
}

func copyFile(src, dest string) error {
	fileContents, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = mkDirForFileIfNeeded(dest)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, fileContents, WRITE_PERMS)
}

func writeFile(path string, contents []byte) error {
	err := mkDirForFileIfNeeded(path)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, contents, WRITE_PERMS)
}

func mkDirForFileIfNeeded(filePath string) error {
	dirPath := path.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, WRITE_PERMS)
	}

	return nil
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}