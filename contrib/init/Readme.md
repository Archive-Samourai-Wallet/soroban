# Setup

On your system (tested on debian)
- copy `soroban` executable into `/usr/local/bin`
- Create user soroban.
- Copy `soroban.service` into `/etc/systemd/system/`
- Reload systemctl: `systemctl daemon-reload`
- Enable soroban: `systemctl enable soroban`
- Start soroban: `service soroban start`
