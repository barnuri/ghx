package update

import "runtime"

func ghxAssetName() string {
	name := "ghx_" + runtime.GOOS + "_" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func ghAssetName(version string) string {
	osNames := map[string]string{"darwin": "macOS", "linux": "linux", "windows": "windows"}
	extension := ".zip"
	if runtime.GOOS == "linux" {
		extension = ".tar.gz"
	}
	return "gh_" + version + "_" + osNames[runtime.GOOS] + "_" + runtime.GOARCH + extension
}

func ghChecksumsName(version string) string {
	return "gh_" + version + "_checksums.txt"
}
