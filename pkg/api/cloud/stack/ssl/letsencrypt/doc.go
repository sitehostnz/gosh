// Package letsencrypt wraps SiteHost's Cloud-Container Let's Encrypt
// endpoints under /cloud/stack/ssl/lets_encrypt. The companion
// service running alongside the shared nginx-proxy on each Cloud
// Container Server provisions certs via the HTTP-01 challenge: as
// long as the stack's hostname (its label / VIRTUAL_HOST env) is
// reachable on port 80, a cert request will succeed.
//
// Endpoints exposed today:
//
//   - List   → cloud/stack/ssl/lets_encrypt/list_all.json   (read)
//   - Create → cloud/stack/ssl/lets_encrypt/create.json     (write, async)
//   - Delete → cloud/stack/ssl/lets_encrypt/delete.json     (write, async)
//   - Renew  → cloud/stack/ssl/lets_encrypt/renew.json      (write, async)
//   - Revoke → cloud/stack/ssl/lets_encrypt/revoke.json     (write, async)
//
// All four write operations return a scheduler job; consumers must
// poll job.Get until state="Completed" before issuing dependent
// calls — see the JobResponse type and the package-level convention
// notes on `pkg/api/cloud/stack`.
package letsencrypt
