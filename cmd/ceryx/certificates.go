package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var errNothingToChange = errors.New("nothing to change; pass at least one flag")

func certificatesCmd(call caller) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificates",
		Aliases: []string{"certs", "cert"},
		Short:   "Manage TLS certificates (domain may be a wildcard like *.example.com)",
		RunE:    group,
	}

	set := &cobra.Command{
		Use:     "set <domain> --cert <file> --key <file>",
		Aliases: []string{"create", "add"},
		Short:   "Create or replace a whole certificate",
		Args:    nargs(1),
		RunE: func(cmd *cobra.Command, a []string) error {
			f := cmd.Flags()
			if !f.Changed("cert") || !f.Changed("key") {
				return usageError{errors.New("--cert and --key are required")}
			}
			body, err := certBody(f)
			if err != nil {
				return err
			}
			body["domain"] = a[0]
			return call("POST", "/api/certificates", body)
		},
	}
	certFlags(set.Flags())

	patch := &cobra.Command{
		Use:   "patch <domain>",
		Short: "Change a stored certificate (exact domain); --cert and --key go together and drop the old chain unless --chain is given",
		Args:  nargs(1),
		RunE: func(cmd *cobra.Command, a []string) error {
			f := cmd.Flags()
			if f.Changed("cert") != f.Changed("key") {
				return usageError{errors.New("--cert and --key must be given together")}
			}
			body, err := certBody(f)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return usageError{errNothingToChange}
			}
			return call("PUT", resource("certificates", a[0]), body)
		},
	}
	certFlags(patch.Flags())

	get := getCmd(call, "certificates", "domain")
	get.Short = "Show the certificate served for a domain (exact, wildcard or route files)"
	cmd.AddCommand(listCmd(call, "/api/certificates"), get, set, patch, deleteCmd(call, "certificates", "domain"))
	return cmd
}

func certFlags(f *pflag.FlagSet) {
	f.String("cert", "", "PEM certificate file (- for stdin)")
	f.String("key", "", "PEM private key file (- for stdin)")
	f.String("chain", "", "PEM intermediate chain file (- for stdin)")
	f.String("meta", "", `metadata JSON object, e.g. '{"vendor":"acme"}'`)
}

func certBody(f *pflag.FlagSet) (map[string]any, error) {
	body := map[string]any{}
	for flag, key := range map[string]string{"cert": "certificate", "key": "private_key", "chain": "chain"} {
		if !f.Changed(flag) {
			continue
		}
		path, _ := f.GetString(flag)
		data, err := readFile(path)
		if err != nil {
			return nil, err
		}
		body[key] = string(data)
	}
	if f.Changed("meta") {
		raw, _ := f.GetString("meta")
		var meta map[string]any
		if err := json.Unmarshal([]byte(raw), &meta); err != nil {
			return nil, usageError{errors.New("--meta must be a JSON object: " + err.Error())}
		}
		body["meta"] = meta
	}
	return body, nil
}

func readFile(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}
