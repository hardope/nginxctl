package certbot

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hardope/nginxctl/internal/system"
)

func IsInstalled() bool {
	_, err := exec.LookPath("certbot")
	return err == nil
}

func Install(osInfo *system.Info) error {
	switch osInfo.Family {
	case system.Debian:
		update := exec.Command("apt-get", "update", "-qq")
		update.Stdout = os.Stdout
		update.Stderr = os.Stderr
		if err := update.Run(); err != nil {
			return fmt.Errorf("apt-get update: %w", err)
		}
		cmd := exec.Command("apt-get", "install", "-y", "certbot", "python3-certbot-nginx")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()

	case system.RHEL:
		// EPEL provides certbot on RHEL/CentOS — ignore error if already present
		epel := exec.Command("dnf", "install", "-y", "epel-release")
		epel.Stdout = os.Stdout
		epel.Stderr = os.Stderr
		_ = epel.Run()

		cmd := exec.Command("dnf", "install", "-y", "certbot", "python3-certbot-nginx")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return fmt.Errorf("unsupported OS family")
}

// Run calls certbot --nginx for the given domains.
// Pass a non-empty email for expiry notifications, or empty to skip registration.
func Run(domains []string, email string) error {
	if len(domains) == 0 {
		return fmt.Errorf("no domains provided")
	}

	args := []string{
		"--nginx",
		"--non-interactive",
		"--agree-tos",
		"--redirect", // auto-add HTTP→HTTPS redirect
	}

	if email != "" {
		args = append(args, "--email", email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	for _, d := range domains {
		args = append(args, "-d", d)
	}

	cmd := exec.Command("certbot", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
