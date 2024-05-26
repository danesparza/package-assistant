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
		// echo -n "$PACKASSIST_GITHUB_GPGKEY" | base64 --decode | gpg --batch --no-tty --passphrase ${PACKASSIST_GITHUB_GPGPASSWORD} --import
	}

	return nil
}
