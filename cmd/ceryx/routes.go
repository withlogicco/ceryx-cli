package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func routesCmd(call caller) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routes",
		Aliases: []string{"route"},
		Short:   "Manage routes (host is a hostname or $wildcard)",
		RunE:    group,
	}

	set := &cobra.Command{
		Use:     "set <source> <target>",
		Aliases: []string{"create", "add"},
		Short:   "Create or replace a whole route; unset settings get server defaults",
		Args:    nargs(2),
		RunE: func(cmd *cobra.Command, a []string) error {
			body := map[string]any{"source": a[0], "target": a[1]}
			if s := routeSettings(cmd.Flags()); len(s) > 0 {
				body["settings"] = s
			}
			return call("POST", "/api/routes", body)
		},
	}
	routeSettingsFlags(set.Flags())

	// Named patch, not update: the API's PUT merges. Kept separate from set (no PUT→POST
	// fallback) so each command is one atomic request and a typo'd host fails with 404.
	patch := &cobra.Command{
		Use:   "patch <host>",
		Short: "Change an existing route; only given flags change, settings are merged",
		Args:  nargs(1),
		RunE: func(cmd *cobra.Command, a []string) error {
			body := map[string]any{}
			if f := cmd.Flags().Lookup("target"); f.Changed {
				body["target"] = f.Value.String()
			}
			if s := routeSettings(cmd.Flags()); len(s) > 0 {
				body["settings"] = s
			}
			if len(body) == 0 {
				return usageError{errNothingToChange}
			}
			return call("PUT", resource("routes", a[0]), body)
		},
	}
	patch.Flags().String("target", "", "upstream to proxy or redirect to")
	routeSettingsFlags(patch.Flags())

	cmd.AddCommand(listCmd(call, "/api/routes"), getCmd(call, "routes", "host"), set, patch, deleteCmd(call, "routes", "host"))
	return cmd
}

func routeSettingsFlags(f *pflag.FlagSet) {
	f.String("mode", "", "proxy or redirect")
	f.Bool("enforce-https", false, "redirect HTTP to HTTPS")
	f.String("certificate-path", "", `certificate file path in the container ("" clears)`)
	f.String("key-path", "", `private key file path in the container ("" clears)`)
	f.Bool("mtls-require-client-certificate", false, "require a client certificate (needs --enforce-https)")
	f.String("mtls-trusted-ca", "", "CA bundle path in the container for client certificates")
	f.Int("mtls-verify-depth", 0, "max intermediates between client certificate and CA")
}

// routeSettings returns only the settings whose flags were given, so the server keeps
// stored values (patch) or applies its defaults (set) for the rest.
func routeSettings(f *pflag.FlagSet) map[string]any {
	s := map[string]any{}
	str := func(name, key string, emptyIsNull bool) {
		if fl := f.Lookup(name); fl.Changed {
			v := fl.Value.String()
			if v == "" && emptyIsNull {
				s[key] = nil
			} else {
				s[key] = v
			}
		}
	}
	str("mode", "mode", false)
	str("certificate-path", "certificate_path", true)
	str("key-path", "key_path", true)
	str("mtls-trusted-ca", "mtls_trusted_ca", false)
	if f.Changed("enforce-https") {
		s["enforce_https"], _ = f.GetBool("enforce-https")
	}
	if f.Changed("mtls-require-client-certificate") {
		s["mtls_require_client_certificate"], _ = f.GetBool("mtls-require-client-certificate")
	}
	if f.Changed("mtls-verify-depth") {
		s["mtls_verify_depth"], _ = f.GetInt("mtls-verify-depth")
	}
	return s
}
