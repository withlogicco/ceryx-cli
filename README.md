# Ceryx CLI

Command line client for the [Ceryx](https://github.com/withlogicco/ceryx) API.

```sh
go install github.com/withlogicco/ceryx-cli/cmd/ceryx@latest
```

The CLI talks to `http://localhost:8080` by default. Point it elsewhere with `--api-url` or `CERYX_API_URL`:

```sh
export CERYX_API_URL=ceryx:8080
```

Responses are the API's JSON on stdout; errors go to stderr. Exit codes: `0` ok, `1` error, `2` usage, `3` not found.

## Routes

### Create a route

Proxy `app.example.com` to the `app` service on port `8000`:

```console
$ ceryx routes set app.example.com app:8000
{
  "settings": {
    "certificate_path": null,
    "enforce_https": false,
    "key_path": null,
    "mode": "proxy",
    "mtls_require_client_certificate": false,
    "mtls_trusted_ca": null,
    "mtls_verify_depth": null
  },
  "source": "app.example.com",
  "target": "app:8000"
}
```

`set` creates the route or replaces it whole: settings you leave out get server defaults.

### Patch a route

Turn on HTTPS enforcement and keep everything else:

```console
$ ceryx routes patch app.example.com --enforce-https
{
  "settings": {
    "certificate_path": null,
    "enforce_https": true,
    "key_path": null,
    "mode": "proxy",
    "mtls_require_client_certificate": false,
    "mtls_trusted_ca": null,
    "mtls_verify_depth": null
  },
  "source": "app.example.com",
  "target": "app:8000"
}
```

`patch` changes only the flags you pass and fails if the route does not exist:

```console
$ ceryx routes patch typo.example.com --enforce-https
error: No route found for typo.example.com (HTTP 404)
$ echo $?
3
```

Pass `--certificate-path ""` or `--key-path ""` to clear a value.

### Create a redirect

```console
$ ceryx routes set old.example.com https://app.example.com --mode redirect
{
  "settings": {
    "certificate_path": null,
    "enforce_https": false,
    "key_path": null,
    "mode": "redirect",
    "mtls_require_client_certificate": false,
    "mtls_trusted_ca": null,
    "mtls_verify_depth": null
  },
  "source": "old.example.com",
  "target": "https://app.example.com"
}
```

### List routes

```console
$ ceryx routes list
[
  {
    "settings": {
      "certificate_path": null,
      "enforce_https": true,
      "key_path": null,
      "mode": "proxy",
      "mtls_require_client_certificate": false,
      "mtls_trusted_ca": null,
      "mtls_verify_depth": null
    },
    "source": "app.example.com",
    "target": "app:8000"
  },
  {
    "settings": {
      "certificate_path": null,
      "enforce_https": false,
      "key_path": null,
      "mode": "redirect",
      "mtls_require_client_certificate": false,
      "mtls_trusted_ca": null,
      "mtls_verify_depth": null
    },
    "source": "old.example.com",
    "target": "https://app.example.com"
  }
]
```

### Delete a route

Deleting prints nothing and succeeds even if the route is already gone:

```console
$ ceryx routes delete old.example.com
$ ceryx routes get old.example.com
error: No route found for old.example.com (HTTP 404)
$ echo $?
3
```

## Certificates

### Upload a certificate

```console
$ ceryx certs set '*.example.com' --cert cert.pem --key key.pem --meta '{"issuer":"internal"}'
{
  "domain": "*.example.com",
  "has_certificate": true,
  "has_chain": false,
  "meta": {
    "issued_at": "2026-11-10T22:13:33Z",
    "issuer": "internal",
    "stored_at": 1794348813
  }
}
```

Add `--chain chain.pem` for intermediates. Any file flag accepts `-` to read stdin.

### See which certificate a domain gets

```console
$ ceryx certs get app.example.com
{
  "domain": "app.example.com",
  "has_certificate": true,
  "has_chain": false,
  "meta": {
    "issued_at": "2026-11-10T22:13:33Z",
    "issuer": "internal",
    "stored_at": 1794348813
  }
}
```

### Patch, list and delete

```console
$ ceryx certs patch '*.example.com' --meta '{"issuer":"acme"}'
$ ceryx certs list
$ ceryx certs delete '*.example.com'
```

`certs patch` changes a stored certificate without uploading it again. `--cert` and `--key` go together, and a new certificate drops the old chain unless you pass `--chain`.

Run `ceryx <command> --help` for all flags.

## Development

Builds and tests run in Docker, so only Docker with buildx is needed:

```sh
docker buildx build --target test .                           # go vet + go test
docker buildx build --platform darwin/arm64 --output bin .    # bin/ceryx for an Apple Silicon Mac
docker buildx build --platform darwin/arm64,linux/amd64 --output dist .  # dist/<os>_<arch>/ceryx
```

In VS Code these are the `build`, `build: release binaries` and `test` tasks; `bin/` is on the `PATH` of VS Code terminals.
