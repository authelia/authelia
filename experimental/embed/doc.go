// Package embed provides tooling useful to embed Authelia into an external go process. This package is considered
// experimental and as such is not supported by the standard versioning policy. It's strongly recommended that care is
// taken when integrating with this package and appropriate tests are conducted when upgrading.
//
// This package and all subpackages are intended to facilitate differing levels of embedability within Authelia. It's
// likely this package and subpackages will break often.
//
// The following considerations should be made in using this package:
//   - It's likely that many methods within this package can panic if not properly utilized.
//   - The package is likely at this stage to be changed abruptly from version to version in a breaking way.
//   - The package will likely have breaking changes at any minor version bump well into the future (breaking changes to
//     this package as a result of changing internal packages will not be a consideration that will slow development).
package embed
