package db

import "github.com/winston-ci/winston/builds"

type DB interface {
	Builds(job string) ([]builds.Build, error)
	CreateBuild(job string) (builds.Build, error)
	GetBuild(job string, id int) (builds.Build, error)

	SaveBuildState(job string, build int, state builds.BuildState) (builds.Build, error)

	BuildLog(job string, build int) ([]byte, error)
	SaveBuildLog(job string, build int, log []byte) error
}