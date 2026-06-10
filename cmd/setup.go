package cmd

import (
	"fmt"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/hardope/nginxctl/internal/certbot"
	nginxpkg "github.com/hardope/nginxctl/internal/nginx"
	"github.com/hardope/nginxctl/internal/system"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Full wizard: install nginx, write reverse-proxy config, optionally add SSL",
	RunE:  runSetup,
}

func runSetup(_ *cobra.Command, _ []string) error {
	requireRoot()

	printBanner("nginxctl setup")

	osInfo, err := system.Detect()
	if err != nil {
		return err
	}
	fmt.Printf("Detected OS: %s\n\n", osInfo.Pretty)

	// ── nginx install ──────────────────────────────────────────────────────
	installed, err := nginxpkg.IsInstalled()
	if err != nil {
		return err
	}

	if !installed {
		var install bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "nginx is not installed. Install it now?",
			Default: true,
		}, &install); err != nil {
			return err
		}
		if !install {
			return fmt.Errorf("nginx is required to continue")
		}
		fmt.Println("Installing nginx...")
		if err := nginxpkg.Install(osInfo); err != nil {
			return fmt.Errorf("install nginx: %w", err)
		}
		if err := nginxpkg.Enable(); err != nil {
			return fmt.Errorf("enable nginx: %w", err)
		}
		fmt.Println("✓ nginx installed and started\n")
	} else {
		fmt.Println("✓ nginx is already installed\n")
	}

	// ── gather proxy config ────────────────────────────────────────────────
	var answers struct {
		Upstream    string
		Domains     string
		MaxBodySize string
	}

	qs := []*survey.Question{
		{
			Name: "upstream",
			Prompt: &survey.Input{
				Message: "Upstream URL:",
				Help:    "e.g. http://localhost:3000  or  https://localhost:8443",
			},
			Validate: survey.Required,
		},
		{
			Name: "domains",
			Prompt: &survey.Input{
				Message: "Domain(s):",
				Help:    "Space-separated — e.g. example.com www.example.com",
			},
			Validate: survey.Required,
		},
		{
			Name: "maxBodySize",
			Prompt: &survey.Input{
				Message: "Max upload size (client_max_body_size):",
				Default: "100M",
				Help:    "e.g. 100M, 1G",
			},
		},
	}

	if err := survey.Ask(qs, &answers); err != nil {
		return err
	}

	domains := strings.Fields(answers.Domains)

	cfg := nginxpkg.Config{
		Domains:     domains,
		Upstream:    answers.Upstream,
		MaxBodySize: answers.MaxBodySize,
		OSInfo:      osInfo,
	}

	content, err := nginxpkg.GenerateConfig(cfg)
	if err != nil {
		return err
	}

	fmt.Println()
	printRule()
	fmt.Print(content)
	printRule()
	fmt.Println()

	var apply bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Apply this configuration?",
		Default: true,
	}, &apply); err != nil {
		return err
	}
	if !apply {
		fmt.Println("Aborted — no changes made.")
		return nil
	}

	// ── write, test, reload ────────────────────────────────────────────────
	if err := nginxpkg.WriteConfig(cfg, content); err != nil {
		return err
	}

	fmt.Println("\nTesting nginx configuration...")
	if err := nginxpkg.Test(); err != nil {
		return err
	}

	if err := nginxpkg.Reload(); err != nil {
		return err
	}
	fmt.Println("✓ nginx configured and reloaded\n")

	// ── optional SSL ───────────────────────────────────────────────────────
	var doSSL bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Set up SSL with Let's Encrypt now?",
		Default: true,
	}, &doSSL); err != nil {
		return err
	}

	if !doSSL {
		fmt.Println("\nDone. Run  sudo nginxctl ssl  later to add SSL.")
		return nil
	}

	return runSSLSetup(domains, osInfo)
}

// runSSLSetup is shared by the setup and ssl commands.
func runSSLSetup(domains []string, osInfo *system.Info) error {
	printBanner("SSL setup")

	fmt.Println("Resolving server public IP...")
	serverIP, err := system.GetPublicIP()
	if err != nil {
		return fmt.Errorf("could not determine server public IP: %w", err)
	}
	fmt.Printf("Server IP: %s\n\n", serverIP)

	printDNSResults(domains, serverIP)

	// ── certbot install ────────────────────────────────────────────────────
	if !certbot.IsInstalled() {
		fmt.Println("certbot is not installed. Installing...")
		if err := certbot.Install(osInfo); err != nil {
			return fmt.Errorf("install certbot: %w", err)
		}
		fmt.Println("✓ certbot installed\n")
	} else {
		fmt.Println("✓ certbot is installed\n")
	}

	var email string
	if err := survey.AskOne(&survey.Input{
		Message: "Email for Let's Encrypt renewal notices (leave blank to skip):",
	}, &email); err != nil {
		return err
	}

	fmt.Println("\nRunning certbot...")
	if err := certbot.Run(domains, email); err != nil {
		return fmt.Errorf("certbot: %w", err)
	}

	fmt.Printf("\n✓ SSL configured — your site is live at https://%s\n", domains[0])
	return nil
}
