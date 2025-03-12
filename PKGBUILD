# Maintainer: Amir Zarrinkafsh <nightah at me dot com>
pkgname=authelia-bin
_pkgname=authelia
pkgver=4.38.19
pkgrel=1
pkgdesc="The Cloud ready multi-factor authentication portal for your Apps. Pre-compiled."
arch=('x86_64' 'aarch64' 'armv7h')
url="https://github.com/authelia/authelia"
license=('Apache-2.0')
provides=('authelia')
conflicts=(
  'authelia'
  'authelia-git'
)
backup=('etc/authelia/configuration.yml')

source_x86_64=("https://github.com/$_pkgname/$_pkgname/releases/download/v$pkgver/$_pkgname-v$pkgver-linux-amd64.tar.gz")
source_aarch64=("https://github.com/$_pkgname/$_pkgname/releases/download/v$pkgver/$_pkgname-v$pkgver-linux-arm64.tar.gz")
source_armv7h=("https://github.com/$_pkgname/$_pkgname/releases/download/v$pkgver/$_pkgname-v$pkgver-linux-arm.tar.gz")

sha256sums_x86_64=('cfefea53308eaff6269f0d4d34bd0f902e8d5822a75488888c4e44445ea876de')
sha256sums_aarch64=('36f9d8cdc318b9e94863908fa38885089daa3a1ba93cfa608f02960686d1b0c4')
sha256sums_armv7h=('5ec505ebbd5e5ab5de5cbca30b0bf1890ace3dbd58a8f4e48fa9aad1cb8f828c')

package() {
  install -Dm644 "$srcdir/$_pkgname.service" "$pkgdir/usr/lib/systemd/system/$_pkgname.service"
  install -Dm644 "$srcdir/$_pkgname.sysusers.conf" "$pkgdir/usr/lib/sysusers.d/$_pkgname.conf"
  install -Dm644 "$srcdir/$_pkgname.tmpfiles" "$pkgdir/usr/lib/tmpfiles.d/$_pkgname.conf"
  install -Dm700 "$srcdir/config.template.yml" "$pkgdir/etc/$_pkgname/configuration.yml"
  if [[ $CARCH == 'x86_64' ]]; then
    install -Dm755 "$srcdir/$_pkgname-linux-amd64" "$pkgdir/usr/bin/$_pkgname"
  elif [[ $CARCH == 'aarch64' ]]; then
    install -Dm755 "$srcdir/$_pkgname-linux-arm64" "$pkgdir/usr/bin/$_pkgname"
  else
    install -Dm755 "$srcdir/$_pkgname-linux-arm" "$pkgdir/usr/bin/$_pkgname"
  fi
}
