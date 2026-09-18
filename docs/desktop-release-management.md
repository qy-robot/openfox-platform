# Desktop release management

RoboCoding stores desktop installers on the platform server and exposes the newest published release through `/downloads.json`. Release drafts and artifacts are not public until an explicit publish request succeeds. Publishing a fourth version removes the oldest published version from the catalog, then retires and deletes its local artifacts; drafts do not count toward the three-version limit.

## Configuration

- `ROBO_RELEASE_STORAGE_DIR`: persistent server-local directory. The default is `data/desktop-releases` relative to the process working directory. Production must mount or back up this directory across deployments.
- `ROBO_RELEASES_INTERNAL_TOKEN`: shared secret used only by the trusted Console BFF. When it is missing, all internal release endpoints fail closed with `503`.
- `ROBO_RELEASE_MAX_FILE_BYTES`: optional decimal byte limit for each installer. The default is `2147483648` (2 GiB). A multipart request is limited to four times this value plus 1 MiB for form metadata.

Only one platform process may write a given server-local release directory. A multi-node deployment must route the internal release API and public artifact requests to the node that owns this directory, or mount one shared filesystem and provide an external single-writer guarantee.

## Internal API

Every request requires `X-Robo-Releases-Token`. Write requests also require `X-Robo-Actor`, containing compact JSON supplied by the authenticated Console BFF:

```json
{"subject":"account-subject","displayName":"Administrator"}
```

Endpoints:

- `GET /v1/internal/releases` returns `{ "success": true, "data": { "latest": Release|null, "releases": Release[] } }`.
- `POST /v1/internal/releases` creates or updates a draft using streamed `multipart/form-data`. Text fields are `version` (semantic version) and `changelog`. File field names are `windows-x64`, `macos-arm64`, `macos-x64`, or `linux-x64`; one to four may be present. Repeated draft uploads for the same version merge platforms and replace only the targets present in the new request.
- `POST /v1/internal/releases/{version}/publish` atomically publishes a draft. Published versions are immutable.

`Release` contains `version`, `changelog`, `status`, `createdAt`, `updatedAt`, nullable `publishedAt`, `actor`, and `artifacts`. Each artifact contains `target`, `fileName`, `sizeBytes`, and `sha256`.

## Public API

- `GET /downloads.json` returns the strict download-page manifest for the newest published release, including `changelog`. With no published release, version, publication time, and changelog are `null` and all targets are unavailable.
- `GET /release-artifacts/{version}/{target}` serves only an artifact that still belongs to one of the retained published versions. Responses use attachment disposition, content sniffing protection, and an immutable cache policy.

The server calculates SHA-256 and byte size from the streamed upload. It writes artifacts and catalog updates through temporary paths and renames them into place. Retention cleanup begins only after the new catalog has been committed, so a failed draft or failed publish does not delete the previously public release.
