# DPI Checkers
🚀 This repository contains checkers that allow you to determine if your “home” ISP has DPI, as well as the specific methods the censor uses for limitations.

## Checkers list
- **RU :: TCP 16-20** => [https://hyperion-cs.github.io/dpi-checkers/ru/tcp-16-20](https://hyperion-cs.github.io/dpi-checkers/ru/tcp-16-20)<br>
Allows to detect TCP 16-20 blocking method in Russia. The tests use publicly available APIs of popular services hosted by providers whose subnets are potentially subject to limitations. The testing process runs right in your browser and the source code is available. VPN should be disabled during the check.<br>

    See [here](https://github.com/net4people/bbs/issues/490) for more details.

    This checker has optional GET parameters:

    | name | type |	default	| description |
    |-|-|-|-|
    | url | string | empty | A custom endpoint to check in addition to the default ones (e.g. your steal-oneself server). Testing custom-testing-page.html should allow cross-origin requests and should provide than 64KB of transferred data. When not specified, the times and provider options are ignored. |
    | times | int | 1 | How many times to access the endpoint in a single HTTP connection (keep-alived). |
    | provider | string | Custom | Provider Name (you can set any name). |
    ||


## Contributing
We would be happy if you could help us improve our checkers through PR or by creating issues.
Also you can star the repository so you don't lose the checkers.
The repository is available [here](https://github.com/hyperion-cs/dpi-checkers).
