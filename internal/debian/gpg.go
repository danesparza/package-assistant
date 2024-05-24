package debian

func InitGPGKey() error {
	//	Does the gpg key exist? This should return stuff
	//	gpg -K

	//	If not, Import the GPG key from environment variables
	//	Note: PACKASSIST_GITHUB_GPGKEY is base64 encoded and in a single line
	// echo -n "$PACKASSIST_GITHUB_GPGKEY" | base64 --decode | gpg --batch --no-tty --passphrase ${PACKASSIST_GITHUB_GPGPASSWORD} --import
	return nil
}
