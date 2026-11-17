// Command ceryx manages routes and certificates of a Ceryx gateway through its API.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// Exit codes are part of the CLI contract; agents branch on them.
const (
	exitError    = 1
	exitUsage    = 2
	exitNotFound = 3
)

type usageError struct{ error }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	root := newRoot(stdout)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.Execute()
	if err == nil {
		return 0
	}
	fmt.Fprintln(stderr, "error:", err)
	var ae *apiError
	var ue usageError
	switch {
	case errors.As(err, &ae) && ae.Status == 404:
		return exitNotFound
	case errors.As(err, &ue):
		return exitUsage
	}
	return exitError
}

func newRoot(stdout io.Writer) *cobra.Command {
	apiURL := os.Getenv("CERYX_API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	var c *client
	root := &cobra.Command{
		Use:   "ceryx",
		Short: "Manage Ceryx routes and certificates",
		Long: `Manage Ceryx routes and certificates through the Ceryx API.

Output is the API's JSON on stdout; errors go to stderr.
Exit codes: 0 ok, 1 error, 2 usage, 3 not found.`,
		Example: `  ceryx routes set app.example.com app:8000 --enforce-https
  ceryx routes patch app.example.com --mode redirect --target https://example.com
  ceryx certs set example.com --cert cert.pem --key key.pem
  CERYX_API_URL=ceryx:8080 ceryx routes list`,
		Version:       version,
		Args:          cobra.ArbitraryArgs, // let group report unknown commands as usage errors
		RunE:          group,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(*cobra.Command, []string) {
			c = newClient(apiURL)
		},
	}
	root.PersistentFlags().StringVar(&apiURL, "api-url", apiURL, "Ceryx API URL (env CERYX_API_URL)")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })
	root.CompletionOptions.DisableDefaultCmd = true

	call := func(method, path string, body any) error {
		data, err := c.do(method, path, body)
		if err != nil {
			return err
		}
		return printJSON(stdout, data)
	}
	root.AddCommand(routesCmd(call), certificatesCmd(call))
	return root
}

type caller func(method, path string, body any) error

// printJSON re-encodes the API response: the server emits keys in random order and
// escapes slashes, so decoding and encoding gives sorted keys and readable URLs.
func printJSON(w io.Writer, data []byte) error {
	if data = bytes.TrimSpace(data); len(data) == 0 {
		return nil
	}
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		_, err = fmt.Fprintf(w, "%s\n", data)
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// group runs command groups. Without it cobra prints help and exits 0 on unknown
// subcommands, hiding typos from agents.
func group(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageError{fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())}
	}
	return cmd.Help()
}

func nargs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(n)(cmd, args); err != nil {
			return usageError{err}
		}
		return nil
	}
}

func listCmd(call caller, path string) *cobra.Command {
	return &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List all", Args: nargs(0),
		RunE: func(*cobra.Command, []string) error { return call("GET", path, nil) },
	}
}

func getCmd(call caller, kind, arg string) *cobra.Command {
	return &cobra.Command{
		Use: "get <" + arg + ">", Short: "Show one", Args: nargs(1),
		RunE: func(_ *cobra.Command, a []string) error { return call("GET", resource(kind, a[0]), nil) },
	}
}

func deleteCmd(call caller, kind, arg string) *cobra.Command {
	return &cobra.Command{
		Use: "delete <" + arg + ">", Aliases: []string{"rm"}, Short: "Delete (succeeds if missing)", Args: nargs(1),
		RunE: func(_ *cobra.Command, a []string) error { return call("DELETE", resource(kind, a[0]), nil) },
	}
}
