# Enable profile picture uploads

Users change their picture from **Settings**, using **Change profile picture**.
The app sends a normalized JPEG to the Go API. Only the server writes to R2.

## Prepare public storage

Pictures are public content. Anyone with a picture URL can read it without
signing in. Do not use this feature for private images.

1. Choose a dedicated R2 bucket with public access already approved for this use.
2. Connect a public HTTPS custom domain to that bucket for production.
   Use the rate-limited `r2.dev` URL only for development.
   Follow Cloudflare's [public bucket guide](https://developers.cloudflare.com/r2/buckets/public-buckets/).
3. Supply server-only S3 access keys scoped to object writes and deletes in that
   bucket. Never put these keys in mobile configuration or `EXPO_PUBLIC_*` variables.

Server-proxied uploads need no R2 browser-upload CORS policy. The S3 API endpoint
is not a public picture origin.

## Configure the server

Set all five variables from `server/.env.example` in the server environment:

- Set `R2_ACCOUNT_ID` to the account's 32-character hexadecimal ID.
- Set `R2_BUCKET` to the bucket name. Use 3-63 lowercase letters, digits, or
  hyphens, with a letter or digit at each end.
- Set `R2_ACCESS_KEY_ID` to the S3 access key ID.
- Set `R2_SECRET_ACCESS_KEY` to its secret. Neither key may contain control characters.
- Set `R2_PUBLIC_BASE_URL` to the public HTTPS origin, such as
  `https://pictures.example.com`. Do not include credentials, a path prefix,
  a query, or a fragment. A trailing slash is accepted.

Leave all five variables empty to disable uploads. Partial or malformed
configuration stops startup. Disabling uploads does not remove stored picture URLs.

The server uses AWS SDK for Go v2 with static keys, region `auto`, and the derived
endpoint `https://<account-id>.r2.cloudflarestorage.com`. Both checksum policies
use `WhenRequired`. There is no endpoint override or default credential lookup.
Cloudflare documents the [Go SDK connection](https://developers.cloudflare.com/r2/examples/aws/aws-sdk-go/).

## Include the schema and native modules

1. Apply `server/db/migrations/00004_profile_picture.sql` through the existing
   Goose migration workflow before starting the updated API.
   Auth queries require the new nullable `users.profile_picture_url` column,
   even when R2 uploads are disabled.
2. Include the updated mobile dependencies and image-picker plugin in a new
   native binary. An over-the-air JavaScript update cannot add native modules.
   The plugin adds the photo-purpose text, not camera or microphone permissions.
3. Keep the public origin available while stored URLs still reference it.
   Changing `R2_PUBLIC_BASE_URL` does not rewrite existing database values.
   Arrange a URL migration separately before retiring an old origin.

## Respect the image limits

Allow one library image per operation. Cancelling the chooser leaves the current
picture unchanged. The app does not request broad library access before selection.

Choose a source within these limits:

- At most 64,000,000 pixels and 16,384 pixels on either side.
- At most 32 MiB when the source size is known.
- Positive, usable dimensions.

The app square-crops and resizes without upscaling to at most 512 by 512 pixels.
It saves JPEG at quality 0.85 and sends at most 2 MiB of bytes.
The app releases native image references before upload and removes only its
generated cache output afterward.

For another API client, send JPEG bytes to `POST /auth/profile-picture` with a
full Bearer access token and `Content-Type: image/jpeg`. Do not send multipart
data, a User ID, an object key, a URL, or storage metadata.

The API accepts at most 2 MiB of actual bytes and dimensions up to 1024 by 1024,
with at most 1,048,576 pixels. It checks JPEG dimensions before full decoding,
then re-encodes at quality 85. This removes EXIF, GPS metadata, and trailing data.
Other image formats and malformed bytes return `validation_failed` with a
`picture` field. Invalid or pending grants return `invalid_session`.
Disabled storage, R2 failures, and reference-save failures return `service_unavailable`.

Read the existing success envelope's `data` as the complete PublicUser.
Its `profile_picture_url` is a stable string, empty when absent.
Mobile maps that field to nullable `profilePictureUrl` in the persisted User.
Sign-in, refresh, and `GET /auth/me` return the same picture field.

## Handle failed or concurrent replacements

Keep each object key immutable. The server generates a fresh UUID-scoped key
under `profile-pictures/<User ID>/<object UUID>.jpg` for every upload.
It records the public URL only after R2 accepts the canonical JPEG.

The repository locks the User row while replacing the reference. After commit,
the server deletes only the actual retired URL under that User's namespace and
the configured public origin. No R2 work holds the database lock.

Treat a failed response after upload as an uncertain save. Do not automatically
retry network or 5xx mutations. Reconcile through `GET /auth/me` or refresh first.
The existing client retries once after a 401 renewal with the same image buffer.

Retain candidate objects on database failures. A failed commit can already be
durable. Failed cleanup can also leave unreferenced objects. Review those objects
against current database references before deleting anything. There is no orphan
cleanup worker or automatic storage lifecycle policy.

Do not treat object deletion as an immediate privacy guarantee. Public caches
can retain retired pictures. Objects carry
`Cache-Control: public, max-age=31536000, immutable`.

If local secure storage fails after a confirmed save, keep the new picture in
memory. Settings reports the local-storage problem without reporting an upload
failure. The next authenticated User read restores the server's picture in memory.

The app writes session updates to SecureStore in call order. A prior picture write
cannot overwrite a later refresh or sign-out.

## Validate before release

Use an approved environment to check chooser access, cropping, cancellation,
sign-out during upload, refresh overlap, concurrent replacement, and public
image serving. Include failure checks for disabled storage, invalid images,
database uncertainty, and local secure-storage errors.

Lint, TypeScript, formatting, and diff checks do not prove those behaviors.
No application, migration, database, R2, or native-build verification was run
for this implementation. The existing HTTP read and write limits remain
15 seconds. R2 upload and delete calls have 10-second and 3-second bounds.
Confirm that the deployed upload path fits those limits before release.

This change does not add camera capture, picture removal, animated pictures,
or pictures to contacts.
