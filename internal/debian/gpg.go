package debian

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strings"
)

func InitGPGKey(ctx context.Context, gpgKey, gpgPassword string) error {
	log.Info().Msg("Initializing GPG key...")
	//	Does the gpg key exist? This should return stuff
	gpgCommand := exec.CommandContext(ctx, "gpg", "-K")
	gpgResponse, err := gpgCommand.Output()
	if err != nil {
		return fmt.Errorf("problem running gpg command: %w", err)
	}

	if len(strings.TrimSpace(string(gpgResponse))) == 0 {
		log.Info().Msg("GPG key does not seem to exist - adding it...")
		//	If not, Import the GPG key from environment variables
		//	Note: PACKASSIST_GITHUB_GPGKEY is base64 encoded and in a single line
		cmd := fmt.Sprintf("echo -n \"%s\" | base64 --decode | gpg --batch --no-tty --passphrase %s --import", gpgKey, gpgPassword)
		_, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			log.Err(err).Msg("problem running gpg command")
		}

		gpgResponse, _ := gpgCommand.Output()
		log.Info().Str("response", string(gpgResponse)).Msg("GPG key list")
	}

	return nil
}

// RefreshPackages refreshes the debian package information in the repo and uses
// the email and gpg password to sign with the gpg key.
func RefreshPackages(ctx context.Context, gpgPassword, gpgEmail, repoFolder string) error {

	log.Info().Str("folder", repoFolder).Str("gpgEmail", gpgEmail).Msg("Refreshing packages...")

	// dpkg-scanpackages --multiversion . > Packages
	cmd := fmt.Sprintf("dpkg-scanpackages --multiversion . > Packages")
	scanPkgCmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	scanPkgCmd.Dir = repoFolder
	_, err := scanPkgCmd.Output()
	if err != nil {
		log.Err(err).Msg("problem running dpkg-scanpackages command")
	}

	// gzip -k -f Packages
	gzipCommand := exec.CommandContext(ctx, "gzip", "-k", "-f", "Packages")
	gzipCommand.Dir = repoFolder
	_, err = gzipCommand.Output()
	if err != nil {
		return fmt.Errorf("problem running gzip command: %w", err)
	}

	// apt-ftparchive release . > Release
	cmd = fmt.Sprintf("apt-ftparchive release . > Release")
	aptCmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	aptCmd.Dir = repoFolder
	_, err = aptCmd.Output()
	if err != nil {
		log.Err(err).Msg("problem running apt-ftparchive command")
	}

	// gpg --pinentry-mode loopback --passphrase ${PASSPHRASE} --batch --no-tty --default-key "${EMAIL}" -abs -o - Release > Release.gpg
	cmd = fmt.Sprintf("gpg --pinentry-mode loopback --passphrase %s --batch --no-tty --default-key \"%s\" -abs -o - Release > Release.gpg", gpgPassword, gpgEmail)
	gpgCmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	gpgCmd.Dir = repoFolder
	_, err = gpgCmd.Output()
	if err != nil {
		log.Err(err).Msg("problem running gpg -abs command")
	}

	// gpg --pinentry-mode loopback --passphrase ${PASSPHRASE} --batch --no-tty --default-key "${EMAIL}" --clearsign -o - Release > InRelease
	cmd = fmt.Sprintf("gpg --pinentry-mode loopback --passphrase %s --batch --no-tty --default-key \"%s\" --clearsign -o - Release > InRelease", gpgPassword, gpgEmail)
	gpg2Cmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	gpg2Cmd.Dir = repoFolder
	_, err = gpg2Cmd.Output()
	if err != nil {
		log.Err(err).Msg("problem running gpg --clearsign command")
	}

	return nil
}
