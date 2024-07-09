package monitor

import (
	"context"
	"github.com/danesparza/package-assistant/internal/cache"
	"github.com/danesparza/package-assistant/internal/debian"
	"github.com/danesparza/package-assistant/internal/repo"
	"github.com/go-redsync/redsync/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Service encapsulates the file versions monitor service
type Service struct {
	StartTime time.Time
	RepoSvc   repo.GitRepoService
}

type FileVersion struct {
	Name    string
	Version string
	Path    string
}

func (service Service) DiscardOldFileVersions(ctx context.Context, rdb *cache.Manager) {
	log.Info().Msg("Starting periodic file versions check...")

	RepoPath := viper.GetString("github.projectfolder")
	gpgPassword := viper.GetString("gpg.password")
	gitName := viper.GetString("git.name")
	gitEmail := viper.GetString("git.email")
	githubUser := viper.GetString("github.user")
	githubPassword := viper.GetString("github.password")

	for {
		select {
		//	Execute it every so often
		case <-time.After(10 * time.Minute):
			//	As we get a request on a channel ...
			log.Debug().Str("projectfolder", viper.GetString("github.projectfolder")).Msg("Checking file versions")

			go func() {
				filesToRemove := FindOldFileVersions(ctx, viper.GetString("github.projectfolder"))
				log.Debug().Strs("filesToRemove", filesToRemove).Msg("Selected files to remove")

				//  Create a mutex and lock
				mutex := rdb.RS.NewMutex(cache.PACKAGE_ASSISTANT_LOCK, redsync.WithExpiry(20*time.Minute))
				if err := mutex.LockContext(ctx); err != nil {
					log.Err(err).Msg("problem getting lock")
					return
				}

				//	Remove the files if we have some to remove
				if len(filesToRemove) > 0 {
					for _, file := range filesToRemove {
						err := os.Remove(file)
						if err != nil {
							log.Err(err).Str("file", file).Msg("problem removing file")
						}
					}

					//	Refresh packages
					err := debian.RefreshPackages(ctx, gpgPassword, gitEmail, RepoPath)
					if err != nil {
						log.Err(err).Msg("problem refreshing packages")
					}

					//	Add all changed / removed files to the commit
					err = service.RepoSvc.AddAll()
					if err != nil {
						log.Err(err).Msg("Error adding files to repo")
					}

					err = service.RepoSvc.CommitAndPush(githubUser, githubPassword, gitName, gitEmail)
					if err != nil {
						log.Err(err).Msg("Error committing files to repo")
					}
				}

				// Release the lock so other processes or threads can obtain a lock.
				if ok, err := mutex.UnlockContext(ctx); !ok || err != nil {
					log.Err(err).Msg("problem releasing lock")
				}

			}() // Launch the goroutine
		case <-ctx.Done():
			log.Info().Msg("File versions check stopping")
			return
		}
	}
}

// FindOldFileVersions finds all versions of all packages in the project folder,
// then returns all file versions older than the 5 most recent for each package.
func FindOldFileVersions(ctx context.Context, RepoPath string) []string {
	retval := make([]string, 0)

	//	First, read the directory:
	files, err := ioutil.ReadDir(RepoPath)
	if err != nil {
		log.Err(err).Str("projectfolder", RepoPath).Msg("Problem reading project folder content")
		return retval
	}

	packageFiles := make(map[string][]FileVersion)
	re := regexp.MustCompile(`^(.*?)-([\d.]+)_.*\.deb$`)

	//	Initial directory scan to gather all file versions and group under their package
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(file.Name())
		if matches == nil {
			continue
		}

		packageName := matches[1]
		version := matches[2]
		filePath := filepath.Join(RepoPath, file.Name())

		packageFiles[packageName] = append(packageFiles[packageName], FileVersion{
			Name:    file.Name(),
			Version: version,
			Path:    filePath,
		})
	}

	for _, files := range packageFiles {
		//	Sort this package's files according to version:
		sort.Slice(files, func(i, j int) bool {
			return versionLess(files[i].Version, files[j].Version)
		})

		//	Assuming we have more than 5 of them,
		//	add everything else to the 'please remove these' results
		if len(files) > 5 {
			for _, file := range files[:len(files)-5] {
				retval = append(retval, file.Path)
			}
		}
	}

	return retval
}

// versionLess compares two version strings and returns true if v1 < v2.
func versionLess(v1, v2 string) bool {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		if v1Parts[i] != v2Parts[i] {
			//	Convert to numbers, then compare
			v1Num, _ := strconv.Atoi(v1Parts[i])
			v2Num, _ := strconv.Atoi(v2Parts[i])
			return v1Num < v2Num
		}
	}

	return len(v1Parts) < len(v2Parts)
}
