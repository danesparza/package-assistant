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
