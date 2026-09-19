# Desktop download manifest

The public download page reads `/downloads.json` at runtime. A deployment can serve this path separately from the app to publish desktop
installers without changing application code. The Go server embeds `web/dist`: if
you use only the embedded static assets, changing the bundled manifest requires
rebuilding the web assets and Go binary.

The manifest must contain exactly one entry for each supported target:

- `windows-x64`
- `macos-arm64`
- `macos-x64`
- `linux-x64`

Set `status` to `available` only after the referenced installer exists. An
available entry requires a non-empty `fileName` and a `url` that is either an
HTTPS URL or a root-relative path such as `/releases/OpenFox.dmg`. An
unavailable entry must use `null` for `url`.

Any manifest with an available entry must also provide both `version` and
`publishedAt`. This prevents a live installer from appearing without a named,
dated release.

`sha256` is optional, but when supplied it must be the full 64-character
hexadecimal SHA-256 digest. `publishedAt` must be an ISO 8601 timestamp. The web
page rejects the complete manifest if required targets are missing, duplicated,
or unsafe, and shows a retryable error instead of exposing an unverified link.

## Publishing

1. Upload each real installer to the website release directory or an HTTPS download host.
2. Verify the download URL, OS/architecture, file name, version, and publication date.
3. Update `downloads.json` and mark only uploaded targets `available`. Keep every
   unpublished target present with `status: unavailable` and `url: null`.
4. Serve `/downloads.json` on the website origin with cache revalidation (for
   example `Cache-Control: no-cache`). If using embedded assets, rebuild web and Go.
5. Open `/download`, verify the available buttons download the intended files, and
   check the unavailable platform states. Cross-origin hosts should send
   `Content-Disposition: attachment` when downloads should save instead of opening.

The checked-in manifest intentionally has no published installers or version.
No production download, installation, signing, or notarization has been verified.
