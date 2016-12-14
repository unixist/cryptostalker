# Systemd unit template

This unit template allows dynamic instances of cryptostalker watching different directories (e.g. when using zfs datasets, different mount points, ...). It is useful on file servers (e.g. ```samba```).

## Installation

Just copy to ```/lib/systemd/system/```

## Usage

Watchout for some bogus systemd escapings (```systemd-escape``` can be useful for testing)!

* ```-``` becomes ```/```

This enables cryptostalker for ```/share/invoice```:

```bash
systemctl enable cryptostalker@-share-invoice.service
systemctl start cryptostalker@-share-invoice.service
```

If your folder already contains```-```, you'll need ```\x2d``` (e.g. ```/internal-share/files/payslip```):

```bash
systemctl enable cryptostalker@-internal\x2dshare-files-payslip.service
systemctl start cryptostalker@-internal\\x2dshare-files-payslip.service # <-- mind the shell-escaped backslash!
```
