# Maintainer: Your Name <your@email.com>
pkgname=ghostgate
gitname=ghostgate
giturl="https://github.com/ghostkellz/ghostgate.git"
pkgver=1.0.0
pkgrel=1
pkgdesc="Modern HTTP server and reverse proxy with ACME, static, and systemd support."
arch=('x86_64')
url="https://github.com/ghostkellz/ghostgate"
license=('AGPL3')
depends=('go')
makedepends=('git')
source=("$giturl")
md5sums=('SKIP')

build() {
  cd "$srcdir/$gitname"
  go build -o ghostgate
}

package() {
  cd "$srcdir/$gitname"
  install -Dm755 ghostgate "$pkgdir/usr/bin/ghostgate"
  install -Dm644 gate.conf "$pkgdir/etc/ghostgate/gate.conf"
  install -Dm644 ghostgate.service "$pkgdir/etc/systemd/system/ghostgate.service"
  cp -r conf.d "$pkgdir/etc/ghostgate/"
  if [ -d static ]; then
    cp -r static "$pkgdir/etc/ghostgate/"
  fi
}
