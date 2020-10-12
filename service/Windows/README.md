# Using nssm to run Galago as a Windows service

## Install

1. Download [nssm (here)](https://nssm.cc/download) and move the `nssm.exe` into the same directory as *this* `README.md`.
2. Right click `install-nssm.bat`, run as administrator.

## Uninstall

1. Right click `uninstall-nssm.bat`, run as administrator.

## Troubleshooting

### Galago can't access images defined in `config/config.yaml`

The service is run as `NT AUTHORITY\NetworkService`.
This user may not have access to other user's private folders.
To fix this, change the user of the server to one that has access to the desired directories. Afterwards restart the service.
