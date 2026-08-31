package server

import "strings"

// Login accounts for the products and image families the API offers.
// The API does not report any of this, so it is recorded here.
const (
	// LoginUserUbuntu is the default account on Ubuntu images on
	// high-performance products.
	LoginUserUbuntu = "ubuntu"

	// LoginUserDebian is the default account on Debian images on
	// high-performance products.
	LoginUserDebian = "debian"

	// LoginUserAlmaLinux is the default account on AlmaLinux images on
	// high-performance products.
	LoginUserAlmaLinux = "almalinux"

	// LoginUserRoot is the account on legacy Xen (LINVPS) products, and
	// in the rescue environment on any product.
	LoginUserRoot = "root"

	// LoginUserWindows is the administrator account on Windows images.
	// Windows is reached over RDP, not SSH, so it is deliberately absent
	// from the lookups below. Untested.
	LoginUserWindows = "Administrator"
)

// Product families, as reported by models.Server.ProductType.
//
// There are three virtual-server tiers, and the naming does not make
// the ordering obvious:
//
//	family    tier                          codes
//	LINVPS    Xen — legacy                  XENLIT, XENPRO, XEN3GB, …
//	SVS       standard performance (KVM)    LSVSP1…LSVSP6, WSVS2…WSVS16
//	HPVS      high performance (KVM)        LHPVS1…LHPVS30
//
// LINVPS is the oldest platform, being migrated away from; prefer HPVS
// for new deployments unless there is a reason not to. See
// [Client.ListProducts] to enumerate the codes a location offers, and
// [Client.CanProvision] to check a specific code against a location.
const (
	// ProductTypeHPVS is the high-performance (KVM) family.
	ProductTypeHPVS = "HPVS"

	// ProductTypeSVS is the standard-performance (KVM) family.
	ProductTypeSVS = "SVS"

	// ProductTypeLINVPS is the legacy Xen family.
	ProductTypeLINVPS = "LINVPS"

	// ProductTypeWINVPS is the Windows family.
	ProductTypeWINVPS = "WINVPS"
)

// hpvsLoginUsers maps a distro family prefix to its login account on
// high-performance products.
var hpvsLoginUsers = map[string]string{
	"ubuntu":    LoginUserUbuntu,
	"debian":    LoginUserDebian,
	"almalinux": LoginUserAlmaLinux,
}

// linvpsLoginUsers maps a distro family prefix to its login account on
// legacy Xen products. Every Linux family observed there uses root, but
// the families are still listed rather than defaulting, so an
// unfamiliar distro returns not-ok instead of a guess.
var linvpsLoginUsers = map[string]string{
	"ubuntu":    LoginUserRoot,
	"debian":    LoginUserRoot,
	"almalinux": LoginUserRoot,
	"centos":    LoginUserRoot,
}

// LoginUserFor returns the account to log in as on a server, given its
// product family and distro — both as reported by [Client.Get] via
// models.Server.ProductType and models.Server.Distro.
//
// # Why this exists
//
// The platform records a login account for every image and uses it when
// provisioning, but no endpoint returns it. Nothing in the API tells a
// caller how to log in to the server it just created, and a wrong guess
// is indistinguishable from a broken key or a firewall problem — which
// makes it an expensive thing to get wrong.
//
// # Both the product family and the distro decide it
//
// Neither alone is enough. The same distro uses a different account on
// each platform:
//
//	product   family   tier                  distro             account
//	LHPVS1    HPVS     high performance      ubuntu-noble       ubuntu
//	XENLIT    LINVPS   Xen (legacy)          ubuntu-noble-pvh   root
//
// So code that assumes the usual cloud-image convention ("ubuntu")
// fails on Xen servers, and code keyed only on the product cannot tell
// a Debian image from an AlmaLinux one on high performance.
//
// # How it was determined
//
// Empirically, in August 2026, at AKLNCT — a public cloud location. One
// server per case, provisioned with an SSH key, then every plausible
// account tried until one authenticated, confirmed with "id -un":
//
//	HPVS    ubuntu-noble       ubuntu
//	HPVS    debian-trixie      debian
//	HPVS    almalinux-10       almalinux
//	LINVPS  ubuntu-noble-pvh   root
//
// Matching is on the distro's leading family name, so ubuntu-jammy and
// ubuntu-focal also resolve, and the mapping holds for releases that did
// not exist when it was written.
//
// # Caveats
//
// Untested, and therefore returning not-ok rather than a guess:
//
//   - **[ProductTypeSVS]** — the standard-performance (KVM) tier was
//     not exercised at all. Do not assume it matches either HPVS or
//     Xen; it is a distinct platform.
//   - **Windows** — reached over RDP as [LoginUserWindows], not SSH,
//     which is why it is absent from the lookups.
//   - **Private Cloud** locations — everything above was public cloud.
//   - **DVS / VDSERV** (developer and desktop families) — not covered.
//
// On LINVPS only one distro was checked; the other Linux families are
// listed on the assumption that root holds across them, which matches
// how those templates are built, but they are inferred rather than
// verified.
//
// The rescue environment is a different image and logs in as
// [LoginUserRoot] whatever the server itself runs; a key injected at
// provision time does not necessarily open it.
//
// ok is false for anything unrecognised rather than guessing. A caller
// that wants to try anyway should say which account it is assuming
// rather than picking one silently.
func LoginUserFor(productType, distro string) (user string, ok bool) {
	switch strings.ToUpper(strings.TrimSpace(productType)) {
	case ProductTypeHPVS:
		return matchDistro(hpvsLoginUsers, distro)
	case ProductTypeLINVPS:
		return matchDistro(linvpsLoginUsers, distro)
	default:
		// Windows, the developer families, and anything unrecognised:
		// no SSH account to offer.
		return "", false
	}
}

// matchDistro resolves a distro against a family map, longest prefix
// winning so a more specific family can be added without reordering.
func matchDistro(families map[string]string, distro string) (string, bool) {
	d := strings.ToLower(strings.TrimSpace(distro))
	if d == "" {
		return "", false
	}
	best, user := "", ""
	for prefix, account := range families {
		if strings.HasPrefix(d, prefix) && len(prefix) > len(best) {
			best, user = prefix, account
		}
	}
	return user, best != ""
}
