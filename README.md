# DPI Checkers
🚀 This repository contains checkers that allow you to determine if your “home” ISP has DPI, as well as the specific methods the censor uses for limitations.

## Checkers list
- **RU :: TCP 16-20** => [https://hyperion-cs.github.io/dpi-checkers/ru/tcp-16-20](https://hyperion-cs.github.io/dpi-checkers/ru/tcp-16-20)<br>
  Allows to detect _TCP 16-20_ blocking method in Russia. The tests use publicly available APIs of popular services hosted by providers whose subnets are potentially subject to limitations. The testing process runs right in your browser and the source code is available. VPN should be disabled during the check.<br>
  This checker has optional GET parameters:
  | name | type |	default	| description |
  |:-:|:-:|:-:|-|
  | timeout | int | `5000` | Timeout for connecting/fetching data from endpoint (in ms). |
  | url | string | — | A custom endpoint to check in addition to the default ones (e.g. your steal-oneself server). The testing endpoint should allow cross-origin requests and provide at least 64KB of data (over the network, including compression, etc.). When not specified, the `times` and `provider` options are ignored. |
  | times | int | `1` | How many times to access the endpoint in a single HTTP connection (_keep-alived_). |
  | provider | string | _Custom_ | Provider name (you can set any name). |

  See [here](https://github.com/net4people/bbs/issues/490) for details on this blocking method.

## Contributing
We would be happy if you could help us improve our checkers through PR or by creating issues.
Also you can star the repository so you don't lose the checkers.
The repository is available [here](https://github.com/hyperion-cs/dpi-checkers).
