<!--
SPDX-FileCopyrightText: 2026 Authelia

SPDX-License-Identifier: Apache-2.0
-->

<p align="center">
  <img src="https://www.authelia.com/images/authelia-title.png" width="350" title="Authelia">
</p>

<p align="center">
  <a href="https://buildkite.com/authelia/authelia"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fbuildkite%2Fd6543d3ece3433f46dbe5fd9fcfaf1f68a6dbc48eb1048bc22%2Fmaster.json&query=%24.message&label=build&logo=buildkite&logoColor=%2314cc80&mode=dark&size=sm&variant=outline"><img alt="Build" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fbuildkite%2Fd6543d3ece3433f46dbe5fd9fcfaf1f68a6dbc48eb1048bc22%2Fmaster.json&query=%24.message&label=build&logo=buildkite&logoColor=%2314cc80&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://codecov.io/gh/authelia/authelia"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fcodecov%2Fc%2Fgithub%2Fauthelia%2Fauthelia.json&query=%24.message&label=coverage&logo=codecov&logoColor=%23f01f7a&mode=dark&size=sm&variant=outline"><img alt="Codecov" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fcodecov%2Fc%2Fgithub%2Fauthelia%2Fauthelia.json&query=%24.message&label=coverage&logo=codecov&logoColor=%23f01f7a&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://bestpractices.coreinfrastructure.org/projects/7128"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/openssf_best_practices-silver.svg?logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA3MCA3MCI%2BPHBhdGggZD0iTTU2Ljk5IDYwLjE1QTMzLjMwIDMzLjMwIDAgMSAxIDU3Ljg0IDExLjQ1TDQ5LjcxIDE5Ljg2QTIxLjYwIDIxLjYwIDAgMSAwIDQ5LjE2IDUxLjQ1WiIgZmlsbD0iIzBmM2I2MyIvPjxwYXRoIGQ9Ik00Ny43NiA0OS44OUExOS41MCAxOS41MCAwIDEgMSA0OC4yNiAyMS4zN0w0MC4xMyAyOS43OUE3LjgwIDcuODAgMCAxIDAgMzkuOTMgNDEuMjBaIiBmaWxsPSIjMzk5NGQwIi8%2BPGNpcmNsZSBjeD0iMzQuNzEiIGN5PSIzNS40IiByPSI1LjciIGZpbGw9IiMwZjNiNjMiLz48L3N2Zz4%3D&logoColor=%233994d0&mode=dark&size=sm&variant=outline"><img alt="OpenSSF Best Practices" src="https://shieldcn.dev/badge/openssf_best_practices-silver.svg?logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA3MCA3MCI%2BPHBhdGggZD0iTTU2Ljk5IDYwLjE1QTMzLjMwIDMzLjMwIDAgMSAxIDU3Ljg0IDExLjQ1TDQ5LjcxIDE5Ljg2QTIxLjYwIDIxLjYwIDAgMSAwIDQ5LjE2IDUxLjQ1WiIgZmlsbD0iIzBmM2I2MyIvPjxwYXRoIGQ9Ik00Ny43NiA0OS44OUExOS41MCAxOS41MCAwIDEgMSA0OC4yNiAyMS4zN0w0MC4xMyAyOS43OUE3LjgwIDcuODAgMCAxIDAgMzkuOTMgNDEuMjBaIiBmaWxsPSIjMzk5NGQwIi8%2BPGNpcmNsZSBjeD0iMzQuNzEiIGN5PSIzNS40IiByPSI1LjciIGZpbGw9IiMwZjNiNjMiLz48L3N2Zz4%3D&logoColor=%230f3b63&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://scorecard.dev/viewer/?uri=github.com/authelia/authelia"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fossf-scorecard%2Fgithub.com%2Fauthelia%2Fauthelia.json&query=%24.message&label=openssf%20scorecard&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyMzAgMjMwIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0xMTgsMiAxMzIsMiAxNDUsNyAxNTYsMTcgMTYzLDM0IDE2MSw1MiAxNDUsNzcgMTM0LDEwNCAxMjksMTIzIDEzNCwxMjUgMTQ2LDEzOSAxNTksMTQxIDE2MywxMzYgMTY2LDEzNiAxODEsMTQ4IDE5NCwxNTMgMjEyLDE1MyAyMTYsMTUxIDIyMSwxNTMgMjI3LDE4NyAyMjUsMjI5IDI3LDIyOSAyNywyMjQgMzUsMjE2IDI0LDE5NCAyNCwxODUgMjcsMTc5IDU0LDE1MCA3MCwxMzkgODMsMTM0IDg2LDEyNSA5OCwxMTggMTAyLDEwNyAxMjAsNzggMTEyLDc1IDg2LDgyIDYxLDgyIDU2LDc4IDU2LDc1IDYzLDY2IDg4LDQ0IDg4LDMxIDk1LDE2IDEwNSw3IDExOCwyWiIvPjwvc3ZnPg%3D%3D&logoColor=%233994d0&mode=dark&size=sm&variant=outline"><img alt="OpenSSF Scorecard" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fossf-scorecard%2Fgithub.com%2Fauthelia%2Fauthelia.json&query=%24.message&label=openssf%20scorecard&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyMzAgMjMwIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0xMTgsMiAxMzIsMiAxNDUsNyAxNTYsMTcgMTYzLDM0IDE2MSw1MiAxNDUsNzcgMTM0LDEwNCAxMjksMTIzIDEzNCwxMjUgMTQ2LDEzOSAxNTksMTQxIDE2MywxMzYgMTY2LDEzNiAxODEsMTQ4IDE5NCwxNTMgMjEyLDE1MyAyMTYsMTUxIDIyMSwxNTMgMjI3LDE4NyAyMjUsMjI5IDI3LDIyOSAyNywyMjQgMzUsMjE2IDI0LDE5NCAyNCwxODUgMjcsMTc5IDU0LDE1MCA3MCwxMzkgODMsMTM0IDg2LDEyNSA5OCwxMTggMTAyLDEwNyAxMjAsNzggMTEyLDc1IDg2LDgyIDYxLDgyIDU2LDc4IDU2LDc1IDYzLDY2IDg4LDQ0IDg4LDMxIDk1LDE2IDEwNSw3IDExOCwyWiIvPjwvc3ZnPg%3D%3D&logoColor=%230f3b63&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://slsa.dev"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/slsa-level_3.svg?logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjMxIDAgMTQwIDE0MCIgZmlsbD0ibm9uZSI%2BPHBhdGggZD0ibTE2MS41MyAzLjA5OTRlLTUgMC4zODUtMC40MzU0My03LjQ5My02LjYyMjItMy4zMTEgMy43NDY1Yy0wLjk4OSAxLjExODQtMS45OTEgMi4yMjIyLTMuMDA4IDMuMzExMWgtMTE3LjF2Ny43OTE5bC02Ljg3OTkgNC4yNDAzIDIuNjIzNCA0LjI1NjVjMS4zNzM1IDIuMjI4NSAyLjc5MjkgNC40MTk2IDQuMjU2NSA2LjU3MjR2OTMuNDIxYy0wLjAzMzkgMWUtMyAtMC4wNjc4IDFlLTMgLTAuMTAxOCAyZS0zbC00Ljk5ODkgMC4xMDIgMC4yMDM1IDkuOTk4IDQuODk3Mi0wLjF2MTMuNzE2aDE0MHYtODguNjg0YzEuNDQtMi4wNzM0IDIuODQtNC4xODMgNC4xOTYtNi4zMjgyIDIuMjc5LTMuNDI5MiAzLjk3MS02LjM2MTMgNS4xMDMtOC40NTc5IDAuNTctMS4wNTQgMC45OTgtMS44OTg2IDEuMjktMi40OTEyIDAuMTQ2LTAuMjk2NCAwLjI1OC0wLjUyOTkgMC4zMzYtMC42OTUzIDAuMDM5LTAuMDgyNyAwLjA2OS0wLjE0ODQgMC4wOTEtMC4xOTY0bDAuMDI3LTAuMDU4NyA5ZS0zIC0wLjAxOTIgM2UtMyAtMC4wMDcxIDFlLTMgLTAuMDAyOSAxZS0zIC0wLjAwMTNjMC02ZS00IDAtMC4wMDExLTQuNTU3LTIuMDU3OGw0LjU1NyAyLjA1NjcgMi4wNTctNC41NTc1LTkuMTE1LTQuMTEzMi0yLjA1NCA0LjU1MTh2MWUtM2wtMWUtMyA5ZS00IC0xZS0zIDAuMDAxN3YyZS0zbC04ZS0zIDAuMDE3NGMtMC4wMTEgMC4wMjM0LTAuMDMgMC4wNjQyLTAuMDU3IDAuMTIxNi0wLjA1NCAwLjExNS0wLjE0MSAwLjI5NjctMC4yNjEgMC41Mzk5LTAuMjM5IDAuNDg2Ni0wLjYxMSAxLjIxODgtMS4xMTYgMi4xNTUtMC4xNTUgMC4yODU5LTAuMzIyIDAuNTkwNy0wLjUwMSAwLjkxMzF2LTMyLjY5em0wIDBoLTEzLjQyN2MtMjIuMDQ2IDIzLjYxOC01MC41OTEgNDAuMjQ2LTgxLjk5MSA0Ny43NzktMTEuODc1LTEwLjU0MS0yMi4zMDUtMjIuODcxLTMwLjg1MS0zNi43MzdsLTIuNjIzNC00LjI1NjUtMS42MzMxIDEuMDA2NXYxNS4wNjljOC43MDc2IDEyLjgwNyAxOC45ODIgMjQuMjU5IDMwLjQ4MiAzNC4xNTYgMTYuNTMgMTQuMjI2IDM1LjU5MSAyNS4yNDIgNTYuMTcgMzIuNDYxLTE3LjQyNCAxMS4zODctMzYuOTYyIDE5LjQ0OC01Ny42MTIgMjMuNjA1LTkuNDc3NCAxLjkwNy0xOS4xOSAyLjk5Mi0yOS4wNCAzLjE5OXYxMC4wMDJsMC4xMDE4LTJlLTNjMTAuNDg0LTAuMjEzIDIwLjgyMy0xLjM2NSAzMC45MTEtMy4zOTYgMjUuNDAzLTUuMTEzIDQ5LjIxNy0xNS43OTYgNjkuNzg2LTMxLjA5IDE1LjAxLTExLjE2MSAyOC4yOTItMjQuNzc5IDM5LjIwMS00MC40OHYtMTguNjI2Yy0wLjk5NiAxLjc5MDgtMi4zOCA0LjEyNy00LjE2MyA2LjgwOGwtMC4wMzMgMC4wNDkxLTAuMDMxIDAuMDQ5OGMtMTAuNTEyIDE2LjYzOS0yMy43NTkgMzEuMDE1LTM4Ljk2MiA0Mi42OC0xOC44ODEtNS43MDktMzYuNTU1LTE0Ljc1OC01Mi4xOC0yNi42NjIgMzEuOTgyLTkuMTI5MiA2MC44MjctMjcuMjUgODIuOTY5LTUyLjMwNHoiIGZpbGw9IiNmMDMxMDAiIGZpbGwtcnVsZT0iZXZlbm9kZCIgY2xpcC1ydWxlPSJldmVub2RkIi8%2BPC9zdmc%2B&logoColor=%23f03100&mode=dark&size=sm&variant=outline"><img alt="SLSA 3" src="https://shieldcn.dev/badge/slsa-level_3.svg?logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjMxIDAgMTQwIDE0MCIgZmlsbD0ibm9uZSI%2BPHBhdGggZD0ibTE2MS41MyAzLjA5OTRlLTUgMC4zODUtMC40MzU0My03LjQ5My02LjYyMjItMy4zMTEgMy43NDY1Yy0wLjk4OSAxLjExODQtMS45OTEgMi4yMjIyLTMuMDA4IDMuMzExMWgtMTE3LjF2Ny43OTE5bC02Ljg3OTkgNC4yNDAzIDIuNjIzNCA0LjI1NjVjMS4zNzM1IDIuMjI4NSAyLjc5MjkgNC40MTk2IDQuMjU2NSA2LjU3MjR2OTMuNDIxYy0wLjAzMzkgMWUtMyAtMC4wNjc4IDFlLTMgLTAuMTAxOCAyZS0zbC00Ljk5ODkgMC4xMDIgMC4yMDM1IDkuOTk4IDQuODk3Mi0wLjF2MTMuNzE2aDE0MHYtODguNjg0YzEuNDQtMi4wNzM0IDIuODQtNC4xODMgNC4xOTYtNi4zMjgyIDIuMjc5LTMuNDI5MiAzLjk3MS02LjM2MTMgNS4xMDMtOC40NTc5IDAuNTctMS4wNTQgMC45OTgtMS44OTg2IDEuMjktMi40OTEyIDAuMTQ2LTAuMjk2NCAwLjI1OC0wLjUyOTkgMC4zMzYtMC42OTUzIDAuMDM5LTAuMDgyNyAwLjA2OS0wLjE0ODQgMC4wOTEtMC4xOTY0bDAuMDI3LTAuMDU4NyA5ZS0zIC0wLjAxOTIgM2UtMyAtMC4wMDcxIDFlLTMgLTAuMDAyOSAxZS0zIC0wLjAwMTNjMC02ZS00IDAtMC4wMDExLTQuNTU3LTIuMDU3OGw0LjU1NyAyLjA1NjcgMi4wNTctNC41NTc1LTkuMTE1LTQuMTEzMi0yLjA1NCA0LjU1MTh2MWUtM2wtMWUtMyA5ZS00IC0xZS0zIDAuMDAxN3YyZS0zbC04ZS0zIDAuMDE3NGMtMC4wMTEgMC4wMjM0LTAuMDMgMC4wNjQyLTAuMDU3IDAuMTIxNi0wLjA1NCAwLjExNS0wLjE0MSAwLjI5NjctMC4yNjEgMC41Mzk5LTAuMjM5IDAuNDg2Ni0wLjYxMSAxLjIxODgtMS4xMTYgMi4xNTUtMC4xNTUgMC4yODU5LTAuMzIyIDAuNTkwNy0wLjUwMSAwLjkxMzF2LTMyLjY5em0wIDBoLTEzLjQyN2MtMjIuMDQ2IDIzLjYxOC01MC41OTEgNDAuMjQ2LTgxLjk5MSA0Ny43NzktMTEuODc1LTEwLjU0MS0yMi4zMDUtMjIuODcxLTMwLjg1MS0zNi43MzdsLTIuNjIzNC00LjI1NjUtMS42MzMxIDEuMDA2NXYxNS4wNjljOC43MDc2IDEyLjgwNyAxOC45ODIgMjQuMjU5IDMwLjQ4MiAzNC4xNTYgMTYuNTMgMTQuMjI2IDM1LjU5MSAyNS4yNDIgNTYuMTcgMzIuNDYxLTE3LjQyNCAxMS4zODctMzYuOTYyIDE5LjQ0OC01Ny42MTIgMjMuNjA1LTkuNDc3NCAxLjkwNy0xOS4xOSAyLjk5Mi0yOS4wNCAzLjE5OXYxMC4wMDJsMC4xMDE4LTJlLTNjMTAuNDg0LTAuMjEzIDIwLjgyMy0xLjM2NSAzMC45MTEtMy4zOTYgMjUuNDAzLTUuMTEzIDQ5LjIxNy0xNS43OTYgNjkuNzg2LTMxLjA5IDE1LjAxLTExLjE2MSAyOC4yOTItMjQuNzc5IDM5LjIwMS00MC40OHYtMTguNjI2Yy0wLjk5NiAxLjc5MDgtMi4zOCA0LjEyNy00LjE2MyA2LjgwOGwtMC4wMzMgMC4wNDkxLTAuMDMxIDAuMDQ5OGMtMTAuNTEyIDE2LjYzOS0yMy43NTkgMzEuMDE1LTM4Ljk2MiA0Mi42OC0xOC44ODEtNS43MDktMzYuNTU1LTE0Ljc1OC01Mi4xOC0yNi42NjIgMzEuOTgyLTkuMTI5MiA2MC44MjctMjcuMjUgODIuOTY5LTUyLjMwNHoiIGZpbGw9IiNmMDMxMDAiIGZpbGwtcnVsZT0iZXZlbm9kZCIgY2xpcC1ydWxlPSJldmVub2RkIi8%2BPC9zdmc%2B&logoColor=%23f03100&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/github/authelia/authelia/license.svg?logo=apache&logoColor=%23d22128&mode=dark&size=sm&variant=outline"><img alt="License" src="https://shieldcn.dev/github/authelia/authelia/license.svg?logo=apache&logoColor=%23d22128&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://opencollective.com/authelia-sponsors"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fopencollective%2Fall%2Fauthelia-sponsors.json&query=%24.message&label=financial%20contributors&logo=opencollective&logoColor=%237fadf2&mode=dark&size=sm&variant=outline"><img alt="Sponsor" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fopencollective%2Fall%2Fauthelia-sponsors.json&query=%24.message&label=financial%20contributors&logo=opencollective&logoColor=%237fadf2&mode=light&size=sm&variant=outline"></picture></a>
</p>

<p align="center">
  <a href="https://github.com/authelia/authelia/releases"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/github/authelia/authelia/release.svg?logo=github&mode=dark&size=sm&variant=outline"><img alt="GitHub Release" src="https://shieldcn.dev/github/authelia/authelia/release.svg?logo=github&logoColor=%23181717&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://hub.docker.com/r/authelia/authelia/tags"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fv%2Fauthelia%2Fauthelia%2Flatest.json%3Fsort%3Dsemver&query=%24.message&label=docker&logo=docker&logoColor=%232496ed&mode=dark&size=sm&variant=outline"><img alt="Docker Tag" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fv%2Fauthelia%2Fauthelia%2Flatest.json%3Fsort%3Dsemver&query=%24.message&label=docker&logo=docker&logoColor=%232496ed&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://hub.docker.com/r/authelia/authelia/tags"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fimage-size%2Fauthelia%2Fauthelia%2Flatest.json%3Fsort%3Dsemver&query=%24.message&label=image%20size&logo=docker&logoColor=%232496ed&mode=dark&size=sm&variant=outline"><img alt="Docker Size" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fimage-size%2Fauthelia%2Fauthelia%2Flatest.json%3Fsort%3Dsemver&query=%24.message&label=image%20size&logo=docker&logoColor=%232496ed&mode=light&size=sm&variant=outline"></picture></a>
  <picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fpulls%2Fauthelia%2Fauthelia.json&query=%24.message&label=pulls&logo=docker&logoColor=%232496ed&mode=dark&size=sm&variant=outline"><img alt="Docker Pulls" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fdocker%2Fpulls%2Fauthelia%2Fauthelia.json&query=%24.message&label=pulls&logo=docker&logoColor=%232496ed&mode=light&size=sm&variant=outline"></picture>
  <a href="https://aur.archlinux.org/packages/authelia/"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Faur%2Fversion%2Fauthelia.json&query=%24.message&label=authelia&logo=archlinux&logoColor=%231793d1&mode=dark&size=sm&variant=outline"><img alt="AUR source version" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Faur%2Fversion%2Fauthelia.json&query=%24.message&label=authelia&logo=archlinux&logoColor=%231793d1&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://aur.archlinux.org/packages/authelia-bin/"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Faur%2Fversion%2Fauthelia-bin.json&query=%24.message&label=authelia-bin&logo=archlinux&logoColor=%231793d1&mode=dark&size=sm&variant=outline"><img alt="AUR binary version" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Faur%2Fversion%2Fauthelia-bin.json&query=%24.message&label=authelia-bin&logo=archlinux&logoColor=%231793d1&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://aur.archlinux.org/packages/authelia-git/"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Faur%2Fversion%2Fauthelia-git.json&query=%24.message&label=authelia-git&logo=archlinux&logoColor=%231793d1&mode=dark&size=sm&variant=outline"><img alt="AUR development version" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Faur%2Fversion%2Fauthelia-git.json&query=%24.message&label=authelia-git&logo=archlinux&logoColor=%231793d1&mode=light&size=sm&variant=outline"></picture></a>
</p>

<p align="center">
  <a href="https://discord.authelia.com"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/discord/707844280412012608.svg?logo=discord&logoColor=%235865f2&mode=dark&size=sm&variant=outline"><img alt="Discord" src="https://shieldcn.dev/discord/707844280412012608.svg?logo=discord&logoColor=%235865f2&mode=light&size=sm&variant=outline"></picture></a>
  <a href="https://matrix.to/#/#support:authelia.com"><picture><source media="(prefers-color-scheme: dark)" srcset="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fmatrix%2Fauthelia-support%3Amatrix.org.json&query=%24.message&label=matrix&logo=matrix&mode=dark&size=sm&variant=outline"><img alt="Matrix" src="https://shieldcn.dev/badge/dynamic/json.svg?url=https%3A%2F%2Fimg.shields.io%2Fmatrix%2Fauthelia-support%3Amatrix.org.json&query=%24.message&label=matrix&logo=matrix&mode=light&size=sm&variant=outline"></picture></a>
</p>

**Authelia** is an open-source authentication and authorization server providing two-factor authentication and single
sign-on (SSO) for your applications via a web portal. It acts as a companion for [reverse proxies](#proxy-support) by
allowing, denying, or redirecting requests.

Documentation is available at [https://www.authelia.com/](https://www.authelia.com/).

The following is a simple diagram of the architecture:

<p align="center" style="margin:50px">
  <img src="https://www.authelia.com/images/archi.png"/>
</p>

**Authelia** can be installed as a standalone service from the [AUR](https://aur.archlinux.org/packages/authelia/),
[APT](https://apt.authelia.com/stable/debian/packages/authelia/),
[FreeBSD Ports](https://svnweb.freebsd.org/ports/head/www/authelia/), or using a
[static binary](https://github.com/authelia/authelia/releases/latest),
[.deb package](https://github.com/authelia/authelia/releases/latest), as a container on [Docker] or [Kubernetes].


Deployment can be orchestrated via the Helm [Chart](https://charts.authelia.com) (beta) leveraging ingress controllers
and ingress configurations.

<p align="center">
  <img src="https://www.authelia.com/images/logos/kubernetes.png" height="100"/>
  <img src="https://www.authelia.com/images/logos/docker.logo.png" width="100">
</p>

Here is what Authelia's portal looks like:

<p align="center">
    <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://www.authelia.com/images/dark.png" width="400">
    <source media="(prefers-color-scheme: light)" srcset="https://www.authelia.com/images/light.png" width="400">
    <img src="https://www.authelia.com/images/light.png" width="400">
  </picture>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://www.authelia.com/images/2fa-methods-dark.png" width="400">
    <source media="(prefers-color-scheme: light)" srcset="https://www.authelia.com/images/2fa-methods-light.png" width="400">
    <img src="https://www.authelia.com/images/2fa-methods-light.png" width="400">
  </picture>
</p>

## Features summary

This is a list of the key features of Authelia:

* [OpenID Connect 1.0 / OAuth 2.0](#openid-connect-10--oauth-20)
* Post-Quantum Cryptography
* Several second factor methods:
  * **[Security Keys](https://www.authelia.com/overview/authentication/security-key/)** that support
    [FIDO2]&nbsp;[WebAuthn] with devices like a [YubiKey].
  * **[Time-based One-Time password](https://www.authelia.com/overview/authentication/one-time-password/)**
    with compatible authenticator applications.
  * **[Mobile Push Notifications](https://www.authelia.com/overview/authentication/push-notification/)**
    with [Duo](https://duo.com/).
* Passwordless Authentication via WebAuthn (Passkeys)
* Password reset with identity verification using email confirmation.
* Access restriction after too many invalid authentication attempts.
* Fine-grained access control using rules which match criteria like subdomain, user, user group membership, request uri,
 request method, and network.
* Choice between one-factor and two-factor policies per-rule.
* Support of basic authentication for endpoints protected by the one-factor policy.
* Highly available using a remote database and Redis as a highly available KV store.
* Compatible with [Traefik](https://doc.traefik.io/traefik) out of the box using the
  [ForwardAuth](https://doc.traefik.io/traefik/middlewares/http/forwardauth/) middleware.
* Curated configuration from [LinuxServer](https://www.linuxserver.io/) via their
  [SWAG](https://docs.linuxserver.io/general/swag) container as well as a
  [guide](https://blog.linuxserver.io/2020/08/26/setting-up-authelia/).
* Compatible with [Caddy] using the [forward_auth](https://caddyserver.com/docs/caddyfile/directives/forward_auth)
  directive.
* Kubernetes Support:
  * Compatible with several Kubernetes Ingress Controllers and Gateways:
    * [ingress-nginx](https://www.authelia.com/integration/kubernetes/nginx-ingress/)
    * [Traefik Kubernetes CRD](https://www.authelia.com/integration/kubernetes/traefik-ingress/#ingressroute)
    * [Traefik Kubernetes Ingress](https://www.authelia.com/integration/kubernetes/traefik-ingress/#ingress)
    * [Istio](https://www.authelia.com/integration/kubernetes/envoy/introduction/)
    * [Envoy Gateway](https://www.authelia.com/integration/kubernetes/envoy/gateway/)
  * Beta support for installing via Helm using our [Charts](https://charts.authelia.com).

For more details take a look at the [Overview](https://www.authelia.com/overview/prologue/introduction/).

If you want to know more about the roadmap, follow [Roadmap](https://www.authelia.com/roadmap).

### OpenID Connect 1.0 / OAuth 2.0

Authelia is [OpenID Certified™] to the Basic OP / Implicit OP / Hybrid OP / Form Post OP / Config OP profiles of the
[OpenID Connect™ protocol]. While this offering is still effectively
[on the roadmap as a beta](https://www.authelia.com/roadmap/active/openid-connect/) it's very comprehensive and well
implemented already, also allowing us comprehensive certification. Read more about the
[OpenID Certified™] status of Authelia in the
[OpenID Connect 1.0 Integration Guide](https://www.authelia.com/integration/openid-connect/introduction/#openid-certified).

<p align="center">
	<a href="https://www.authelia.com/integration/openid-connect/introduction/#openid-certified" target="_blank">
		<picture>
			<img src="https://www.authelia.com/images/oid-certification.jpg" width="400" title="OpenID Certified™ by Authelia to the Basic OP / Implicit OP / Hybrid OP / Form Post OP / Config OP of the OpenID Connect™ protocol">
		</picture>
	</a>
</p>

## Proxy support

Authelia works in combination with [nginx], [Traefik], [Caddy], [Skipper], [Envoy], or [HAProxy].

<p align="center">
  <img src="https://www.authelia.com/images/logos/nginx.png" height="50"/>
  <img src="https://www.authelia.com/images/logos/traefik.png" height="50"/>
  <img src="https://www.authelia.com/images/logos/caddy.png" height="50"/>
  <img src="https://www.authelia.com/images/logos/envoy.png" height="50"/>
  <img src="https://www.authelia.com/images/logos/haproxy.png" height="50"/>
</p>

## Getting Started

See the [Get Started Guide](https://www.authelia.com/integration/prologue/get-started/) or one of the curated examples
below.

### docker compose

The `docker compose` bundles act as a starting point for anyone wanting to see Authelia in action. You will have to
customize them to your needs as they come with self-signed certificates.

#### [Local](https://www.authelia.com/integration/deployment/docker/#local)
The Local compose bundle is intended to test Authelia without worrying about configuration.
It's meant to be used for scenarios where the server is not be exposed to the internet.
Domains will be defined in the local hosts file and self-signed certificates will be utilized.

#### [Lite](https://www.authelia.com/integration/deployment/docker/#lite)
The Lite compose bundle is intended for scenarios where the server will be exposed to the internet, domains and DNS will
need to be setup accordingly and certificates will be generated through LetsEncrypt. The Lite element refers to minimal
external dependencies; File based user storage, SQLite based configuration storage. In this configuration, the service
will not scale well.

## Deployment

Now that you have tested **Authelia** and you want to try it out in your own infrastructure,
you can learn how to deploy and use it with [Deployment](https://www.authelia.com/integration/deployment/introduction/).
This guide will show you how to deploy it on bare metal as well as on
[Kubernetes](https://kubernetes.io/).

## Security

Authelia takes security very seriously. If you discover a vulnerability in Authelia, please see our
[Security Policy](https://github.com/authelia/authelia/security/policy).

For more information about [security](https://www.authelia.com/policies/security/) related matters, please read
[the documentation](https://www.authelia.com/policies/security/).

## Contact Options

Several contact options exist for our community, the primary one being [Matrix](#matrix). These are in addition to
[GitHub issues](https://github.com/authelia/authelia/issues) for creating a
[new issue](https://github.com/authelia/authelia/issues/new/choose).

### Matrix

Community members are invited to join the [Matrix Space](https://matrix.to/#/#community:authelia.com) which includes
both the [Support Room](https://matrix.to/#/#support:authelia.com) and the
[Contributing Room](https://matrix.to/#/#contributing:authelia.com).

- The core team members are identified as administrators in the Space and individual Rooms.
- All channels are linked to [Discord](#discord).

### Discord

Community members are invited to join the [Discord Server](https://discord.authelia.com).

- The core team members are identified by the <span style="color:#BA55D3;">**CORE TEAM**</span> role in Discord.
- The [#support] and [#contributing] channels are linked to [Matrix](#matrix).

### Email

You can contact the core team by email via [team@authelia.com](mailto:team@authelia.com). Please note the
[security@authelia.com](mailto:security@authelia.com) is also available but is strictly reserved for [security] related
matters.

## Breaking changes

Since Authelia is still under active development, it is subject to breaking changes. It's recommended to pin a version
tag instead of using the `latest` tag and reading the [release notes](https://github.com/authelia/authelia/releases)
before upgrading. This is where you will find information about breaking changes and what you should do to overcome
said changes.

## Why Open Source?

You might wonder why Authelia is open source while it adds a great deal of security and user experience to your
infrastructure at zero cost. It is open source because we firmly believe that security should be available for all to
benefit in the face of the battlefield which is the Internet, with near zero effort.

Additionally, keeping the code open source is a way to leave it auditable by anyone who is willing to contribute. This
way, you can be confident that the product remains secure and does not act maliciously.

It's important to keep in mind Authelia is not directly exposed on the
Internet (your reverse proxies are) however, it's still the control plane for your internal security so take care of it!

## Contribute

If you want to contribute to Authelia, please read our [contribution guidelines](CONTRIBUTING.md).

Authelia exists thanks to all the people who contribute so don't be shy, come chat with us on either [Matrix](#matrix)
or [Discord](#discord) and start contributing too.

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/clems4ever"><img src="https://avatars.githubusercontent.com/u/3193257?v=4?s=100" width="100px;" alt="Clément Michaud"/><br /><sub><b>Clément Michaud</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=clems4ever" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=clems4ever" title="Documentation">📖</a> <a href="#ideas-clems4ever" title="Ideas, Planning, & Feedback">🤔</a> <a href="#maintenance-clems4ever" title="Maintenance">🚧</a> <a href="#question-clems4ever" title="Answering Questions">💬</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Aclems4ever" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/authelia/authelia/commits?author=clems4ever" title="Tests">⚠️</a> <a href="#mentoring-clems4ever" title="Mentoring">🧑‍🏫</a> <a href="#infra-clems4ever" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#design-clems4ever" title="Design">🎨</a> <a href="#userTesting-clems4ever" title="User Testing">📓</a> <a href="#tool-clems4ever" title="Tools">🔧</a> <a href="#research-clems4ever" title="Research">🔬</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/nightah"><img src="https://avatars.githubusercontent.com/u/3339418?v=4?s=100" width="100px;" alt="Amir Zarrinkafsh"/><br /><sub><b>Amir Zarrinkafsh</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=nightah" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=nightah" title="Documentation">📖</a> <a href="#ideas-nightah" title="Ideas, Planning, & Feedback">🤔</a> <a href="#maintenance-nightah" title="Maintenance">🚧</a> <a href="#question-nightah" title="Answering Questions">💬</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Anightah" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/authelia/authelia/commits?author=nightah" title="Tests">⚠️</a> <a href="#mentoring-nightah" title="Mentoring">🧑‍🏫</a> <a href="#infra-nightah" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#design-nightah" title="Design">🎨</a> <a href="#userTesting-nightah" title="User Testing">📓</a> <a href="#tool-nightah" title="Tools">🔧</a> <a href="#research-nightah" title="Research">🔬</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/james-d-elliott"><img src="https://avatars.githubusercontent.com/u/3903683?v=4?s=100" width="100px;" alt="James Elliott"/><br /><sub><b>James Elliott</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=james-d-elliott" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=james-d-elliott" title="Documentation">📖</a> <a href="#ideas-james-d-elliott" title="Ideas, Planning, & Feedback">🤔</a> <a href="#maintenance-james-d-elliott" title="Maintenance">🚧</a> <a href="#question-james-d-elliott" title="Answering Questions">💬</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Ajames-d-elliott" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/authelia/authelia/commits?author=james-d-elliott" title="Tests">⚠️</a> <a href="#mentoring-james-d-elliott" title="Mentoring">🧑‍🏫</a> <a href="#infra-james-d-elliott" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#design-james-d-elliott" title="Design">🎨</a> <a href="#userTesting-james-d-elliott" title="User Testing">📓</a> <a href="#tool-james-d-elliott" title="Tools">🔧</a> <a href="#research-james-d-elliott" title="Research">🔬</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/n4kre"><img src="https://avatars.githubusercontent.com/u/14371127?v=4?s=100" width="100px;" alt="Antoine Favre"/><br /><sub><b>Antoine Favre</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3An4kre" title="Bug reports">🐛</a> <a href="#ideas-n4kre" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/BankaiNoJutsu"><img src="https://avatars.githubusercontent.com/u/2241519?v=4?s=100" width="100px;" alt="BankaiNoJutsu"/><br /><sub><b>BankaiNoJutsu</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=BankaiNoJutsu" title="Code">💻</a> <a href="#design-BankaiNoJutsu" title="Design">🎨</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/p-rintz"><img src="https://avatars.githubusercontent.com/u/13933258?v=4?s=100" width="100px;" alt="Philipp Rintz"/><br /><sub><b>Philipp Rintz</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=p-rintz" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://callanbryant.co.uk/"><img src="https://avatars.githubusercontent.com/u/208440?v=4?s=100" width="100px;" alt="Callan Bryant"/><br /><sub><b>Callan Bryant</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=naggie" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=naggie" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ViViDboarder"><img src="https://avatars.githubusercontent.com/u/137025?v=4?s=100" width="100px;" alt="Ian"/><br /><sub><b>Ian</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ViViDboarder" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/FrozenDragoon"><img src="https://avatars.githubusercontent.com/u/5301673?v=4?s=100" width="100px;" alt="FrozenDragoon"/><br /><sub><b>FrozenDragoon</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=FrozenDragoon" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/vdot0x23"><img src="https://avatars.githubusercontent.com/u/40716069?v=4?s=100" width="100px;" alt="vdot0x23"/><br /><sub><b>vdot0x23</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=vdot0x23" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/alexw1982"><img src="https://avatars.githubusercontent.com/u/11628284?v=4?s=100" width="100px;" alt="alexw1982"/><br /><sub><b>alexw1982</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=alexw1982" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Sohalt"><img src="https://avatars.githubusercontent.com/u/2157287?v=4?s=100" width="100px;" alt="Sohalt"/><br /><sub><b>Sohalt</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Sohalt" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=Sohalt" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Tedyst"><img src="https://avatars.githubusercontent.com/u/13637623?v=4?s=100" width="100px;" alt="Stoica Tedy"/><br /><sub><b>Stoica Tedy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Tedyst" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Chemsmith"><img src="https://avatars.githubusercontent.com/u/9061024?v=4?s=100" width="100px;" alt="Dylan Smith"/><br /><sub><b>Dylan Smith</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Chemsmith" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/LukasK13"><img src="https://avatars.githubusercontent.com/u/24586740?v=4?s=100" width="100px;" alt="Lukas Klass"/><br /><sub><b>Lukas Klass</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=LukasK13" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://staiger.it/"><img src="https://avatars.githubusercontent.com/u/9325003?v=4?s=100" width="100px;" alt="Philipp Staiger"/><br /><sub><b>Philipp Staiger</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=lippl" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=lippl" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=lippl" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://yaleman.org/"><img src="https://avatars.githubusercontent.com/u/168188?v=4?s=100" width="100px;" alt="James Hodgkinson"/><br /><sub><b>James Hodgkinson</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=yaleman" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://chris.smith.xyz/"><img src="https://avatars.githubusercontent.com/u/1979423?v=4?s=100" width="100px;" alt="Chris Smith"/><br /><sub><b>Chris Smith</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=chris13524" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/mqmq0"><img src="https://avatars.githubusercontent.com/u/13240971?v=4?s=100" width="100px;" alt="Mihály"/><br /><sub><b>Mihály</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=mqmq0" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://iret.xyz/"><img src="https://avatars.githubusercontent.com/u/6560655?v=4?s=100" width="100px;" alt="Silver Bullet"/><br /><sub><b>Silver Bullet</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=SilverBut" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/skenmy"><img src="https://avatars.githubusercontent.com/u/1454505?v=4?s=100" width="100px;" alt="Paul Williams"/><br /><sub><b>Paul Williams</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=skenmy" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=skenmy" title="Tests">⚠️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ntimo"><img src="https://avatars.githubusercontent.com/u/6145026?v=4?s=100" width="100px;" alt="Timo"/><br /><sub><b>Timo</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ntimo" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/andrewkliskey"><img src="https://avatars.githubusercontent.com/u/44645768?v=4?s=100" width="100px;" alt="Andrew Kliskey"/><br /><sub><b>Andrew Kliskey</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=andrewkliskey" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://kristofmattei.be/"><img src="https://avatars.githubusercontent.com/u/864376?v=4?s=100" width="100px;" alt="Kristof Mattei"/><br /><sub><b>Kristof Mattei</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Kristof-Mattei" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.zmiguel.me/"><img src="https://avatars.githubusercontent.com/u/4400540?v=4?s=100" width="100px;" alt="ZMiguel Valdiviesso"/><br /><sub><b>ZMiguel Valdiviesso</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=zmiguel" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/akusei"><img src="https://avatars.githubusercontent.com/u/12972900?v=4?s=100" width="100px;" alt="akusei"/><br /><sub><b>akusei</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=akusei" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=akusei" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Peaches491"><img src="https://avatars.githubusercontent.com/u/494334?v=4?s=100" width="100px;" alt="Daniel Miller"/><br /><sub><b>Daniel Miller</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Peaches491" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/dustins"><img src="https://avatars.githubusercontent.com/u/14645?v=4?s=100" width="100px;" alt="Dustin Sweigart"/><br /><sub><b>Dustin Sweigart</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dustins" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=dustins" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=dustins" title="Tests">⚠️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/rogue780"><img src="https://avatars.githubusercontent.com/u/247716?v=4?s=100" width="100px;" alt="Shawn Haggard"/><br /><sub><b>Shawn Haggard</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=rogue780" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=rogue780" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/kevynb"><img src="https://avatars.githubusercontent.com/u/4941215?v=4?s=100" width="100px;" alt="Kevyn Bruyere"/><br /><sub><b>Kevyn Bruyere</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=kevynb" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ducksecops"><img src="https://avatars.githubusercontent.com/u/25612094?v=4?s=100" width="100px;" alt="Daniel Sutton"/><br /><sub><b>Daniel Sutton</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ducksecops" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://www.xenuser.org/"><img src="https://avatars.githubusercontent.com/u/2216868?v=4?s=100" width="100px;" alt="Valentin Höbel"/><br /><sub><b>Valentin Höbel</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=xenuser" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/thehedgefrog"><img src="https://avatars.githubusercontent.com/u/38590447?v=4?s=100" width="100px;" alt="thehedgefrog"/><br /><sub><b>thehedgefrog</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=thehedgefrog" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ViRb3"><img src="https://avatars.githubusercontent.com/u/2650170?v=4?s=100" width="100px;" alt="Victor"/><br /><sub><b>Victor</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ViRb3" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/whiskerch"><img src="https://avatars.githubusercontent.com/u/35109315?v=4?s=100" width="100px;" alt="Chris Whisker"/><br /><sub><b>Chris Whisker</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=whiskerch" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/nasatome"><img src="https://avatars.githubusercontent.com/u/18271791?v=4?s=100" width="100px;" alt="nasatome"/><br /><sub><b>nasatome</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=nasatome" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/bbros-dev"><img src="https://avatars.githubusercontent.com/u/60454087?v=4?s=100" width="100px;" alt="Begley Brothers (Development)"/><br /><sub><b>Begley Brothers (Development)</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=bbros-dev" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://mikekusold.com/"><img src="https://avatars.githubusercontent.com/u/509966?v=4?s=100" width="100px;" alt="Mike Kusold"/><br /><sub><b>Mike Kusold</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=kusold" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://dzervas.gr/"><img src="https://avatars.githubusercontent.com/u/1029195?v=4?s=100" width="100px;" alt="Dimitris Zervas"/><br /><sub><b>Dimitris Zervas</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dzervas" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://paypal.me/DHoung"><img src="https://avatars.githubusercontent.com/u/52870424?v=4?s=100" width="100px;" alt="TheCatLady"/><br /><sub><b>TheCatLady</b></sub></a><br /><a href="#ideas-TheCatLady" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://lauri.vosandi.com/"><img src="https://avatars.githubusercontent.com/u/194685?v=4?s=100" width="100px;" alt="Lauri Võsandi"/><br /><sub><b>Lauri Võsandi</b></sub></a><br /><a href="#ideas-laurivosandi" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/knnnrd"><img src="https://avatars.githubusercontent.com/u/5852381?v=4?s=100" width="100px;" alt="Kennard Vermeiren"/><br /><sub><b>Kennard Vermeiren</b></sub></a><br /><a href="#ideas-knnnrd" title="Ideas, Planning, & Feedback">🤔</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ThinkChaos"><img src="https://avatars.githubusercontent.com/u/4761135?v=4?s=100" width="100px;" alt="ThinkChaos"/><br /><sub><b>ThinkChaos</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ThinkChaos" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=ThinkChaos" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=ThinkChaos" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/except"><img src="https://avatars.githubusercontent.com/u/26675576?v=4?s=100" width="100px;" alt="Hasan"/><br /><sub><b>Hasan</b></sub></a><br /><a href="#security-except" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://blog.dchidell.com"><img src="https://avatars.githubusercontent.com/u/26146619?v=4?s=100" width="100px;" alt="David Chidell"/><br /><sub><b>David Chidell</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dchidell" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/mardom1"><img src="https://avatars.githubusercontent.com/u/32371724?v=4?s=100" width="100px;" alt="Marcel Marquardt"/><br /><sub><b>Marcel Marquardt</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Amardom1" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://cdine.org"><img src="https://avatars.githubusercontent.com/u/127512?v=4?s=100" width="100px;" alt="Ian Gallagher"/><br /><sub><b>Ian Gallagher</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=craSH" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://wuhanstudio.cc"><img src="https://avatars.githubusercontent.com/u/15157070?v=4?s=100" width="100px;" alt="Wu Han"/><br /><sub><b>Wu Han</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=wuhanstudio" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/lavih"><img src="https://avatars.githubusercontent.com/u/47455309?v=4?s=100" width="100px;" alt="lavih"/><br /><sub><b>lavih</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=lavih" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="http://jonbayl"><img src="https://avatars.githubusercontent.com/u/30201351?v=4?s=100" width="100px;" alt="Jon B. "/><br /><sub><b>Jon B. </b></sub></a><br /><a href="#security-jonbayl" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/AlexGustafsson"><img src="https://avatars.githubusercontent.com/u/14974112?v=4?s=100" width="100px;" alt="Alex Gustafsson"/><br /><sub><b>Alex Gustafsson</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=AlexGustafsson" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=AlexGustafsson" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.aarsen.me/"><img src="https://avatars.githubusercontent.com/u/7805050?v=4?s=100" width="100px;" alt="Arsenović Arsen"/><br /><sub><b>Arsenović Arsen</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ArsenArsen" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=ArsenArsen" title="Tests">⚠️</a> <a href="#security-ArsenArsen" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/dakriy"><img src="https://avatars.githubusercontent.com/u/13756065?v=4?s=100" width="100px;" alt="dakriy"/><br /><sub><b>dakriy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=dakriy" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/davama"><img src="https://avatars.githubusercontent.com/u/5359152?v=4?s=100" width="100px;" alt="Dave"/><br /><sub><b>Dave</b></sub></a><br /><a href="#userTesting-davama" title="User Testing">📓</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/nreymundo"><img src="https://avatars.githubusercontent.com/u/5833447?v=4?s=100" width="100px;" alt="Nicolas Reymundo"/><br /><sub><b>Nicolas Reymundo</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=nreymundo" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/polandy"><img src="https://avatars.githubusercontent.com/u/3670670?v=4?s=100" width="100px;" alt="polandy"/><br /><sub><b>polandy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=polandy" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/you1996"><img src="https://avatars.githubusercontent.com/u/45292366?v=4?s=100" width="100px;" alt="yossbg"/><br /><sub><b>yossbg</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=you1996" title="Code">💻</a> <a href="#design-you1996" title="Design">🎨</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/mpdcampbell"><img src="https://avatars.githubusercontent.com/u/47434940?v=4?s=100" width="100px;" alt="Michael Campbell"/><br /><sub><b>Michael Campbell</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=mpdcampbell" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://sievenpiper.co"><img src="https://avatars.githubusercontent.com/u/1131882?v=4?s=100" width="100px;" alt="Justin Sievenpiper"/><br /><sub><b>Justin Sievenpiper</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=jsievenpiper" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/kaysond"><img src="https://avatars.githubusercontent.com/u/1147328?v=4?s=100" width="100px;" alt="Aram Akhavan"/><br /><sub><b>Aram Akhavan</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=kaysond" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://skhuf.net"><img src="https://avatars.githubusercontent.com/u/286341?v=4?s=100" width="100px;" alt="Shadow"/><br /><sub><b>Shadow</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=shadow7412" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tarioch"><img src="https://avatars.githubusercontent.com/u/2998148?v=4?s=100" width="100px;" alt="Patrick Ruckstuhl"/><br /><sub><b>Patrick Ruckstuhl</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=tarioch" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/FineWolf"><img src="https://avatars.githubusercontent.com/u/203591?v=4?s=100" width="100px;" alt="Andrew Moore"/><br /><sub><b>Andrew Moore</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=FineWolf" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=FineWolf" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/commits?author=FineWolf" title="Tests">⚠️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="http://www.dennisgaida.de"><img src="https://avatars.githubusercontent.com/u/2392217?v=4?s=100" width="100px;" alt="Dennis Gaida"/><br /><sub><b>Dennis Gaida</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=DennisGaida" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Alestrix"><img src="https://avatars.githubusercontent.com/u/7452860?v=4?s=100" width="100px;" alt="Alestrix"/><br /><sub><b>Alestrix</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Alestrix" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/bgh-github"><img src="https://avatars.githubusercontent.com/u/99472455?v=4?s=100" width="100px;" alt="bgh-github"/><br /><sub><b>bgh-github</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=bgh-github" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/mind-ar"><img src="https://avatars.githubusercontent.com/u/10672208?v=4?s=100" width="100px;" alt="Manuel Nuñez"/><br /><sub><b>Manuel Nuñez</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=mind-ar" title="Code">💻</a> <a href="#translation-mind-ar" title="Translation">🌍</a> <a href="https://github.com/authelia/authelia/commits?author=mind-ar" title="Documentation">📖</a> <a href="https://github.com/authelia/authelia/issues?q=author%3Amind-ar" title="Bug reports">🐛</a> <a href="#design-mind-ar" title="Design">🎨</a> <a href="https://github.com/authelia/authelia/commits?author=mind-ar" title="Tests">⚠️</a> <a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3Amind-ar" title="Reviewed Pull Requests">👀</a> <a href="#research-mind-ar" title="Research">🔬</a> <a href="#ideas-mind-ar" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/protvis74"><img src="https://avatars.githubusercontent.com/u/50554836?v=4?s=100" width="100px;" alt="protvis74"/><br /><sub><b>protvis74</b></sub></a><br /><a href="#translation-protvis74" title="Translation">🌍</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://itjamie.com"><img src="https://avatars.githubusercontent.com/u/1613241?v=4?s=100" width="100px;" alt="Jamie (Bear) Murphy "/><br /><sub><b>Jamie (Bear) Murphy </b></sub></a><br /><a href="https://github.com/authelia/authelia/pulls?q=is%3Apr+reviewed-by%3AITJamie" title="Reviewed Pull Requests">👀</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Beanow"><img src="https://avatars.githubusercontent.com/u/497556?v=4?s=100" width="100px;" alt="Robin van Boven"/><br /><sub><b>Robin van Boven</b></sub></a><br /><a href="#security-Beanow" title="Security">🛡️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="http://www.cybertrol.com"><img src="https://avatars.githubusercontent.com/u/1178293?v=4?s=100" width="100px;" alt="alphabet5"/><br /><sub><b>alphabet5</b></sub></a><br /><a href="#ideas-alphabet5" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/rjmidau"><img src="https://avatars.githubusercontent.com/u/8134995?v=4?s=100" width="100px;" alt="Robert Meredith"/><br /><sub><b>Robert Meredith</b></sub></a><br /><a href="#ideas-rjmidau" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/adriang-90"><img src="https://avatars.githubusercontent.com/u/60886162?v=4?s=100" width="100px;" alt="Adrian Gąsior"/><br /><sub><b>Adrian Gąsior</b></sub></a><br /><a href="#security-adriang-90" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://jamesw.link/me"><img src="https://avatars.githubusercontent.com/u/8067792?v=4?s=100" width="100px;" alt="James White"/><br /><sub><b>James White</b></sub></a><br /><a href="#question-jamesmacwhite" title="Answering Questions">💬</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.zxlim.xyz"><img src="https://avatars.githubusercontent.com/u/19372079?v=4?s=100" width="100px;" alt="Zhao Xiang Lim"/><br /><sub><b>Zhao Xiang Lim</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=zxlim" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Auzborn123"><img src="https://avatars.githubusercontent.com/u/42992103?v=4?s=100" width="100px;" alt="Auzborn123"/><br /><sub><b>Auzborn123</b></sub></a><br /><a href="#translation-Auzborn123" title="Translation">🌍</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/SvanGlan"><img src="https://avatars.githubusercontent.com/u/106152205?v=4?s=100" width="100px;" alt="SvanGlan"/><br /><sub><b>SvanGlan</b></sub></a><br /><a href="#translation-SvanGlan" title="Translation">🌍</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/HannesJo0139"><img src="https://avatars.githubusercontent.com/u/42114183?v=4?s=100" width="100px;" alt="HannesJo0139"/><br /><sub><b>HannesJo0139</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=HannesJo0139" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/andreas-berg"><img src="https://avatars.githubusercontent.com/u/39428693?v=4?s=100" width="100px;" alt="andreas-berg"/><br /><sub><b>andreas-berg</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Aandreas-berg" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://radenac.me"><img src="https://avatars.githubusercontent.com/u/47008408?v=4?s=100" width="100px;" alt="Clément Radenac"/><br /><sub><b>Clément Radenac</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=clem3109" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/boomam"><img src="https://avatars.githubusercontent.com/u/37086258?v=4?s=100" width="100px;" alt="boomam"/><br /><sub><b>boomam</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=boomam" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Northguy"><img src="https://avatars.githubusercontent.com/u/1189058?v=4?s=100" width="100px;" alt="Northguy"/><br /><sub><b>Northguy</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Northguy" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/polarathene"><img src="https://avatars.githubusercontent.com/u/5098581?v=4?s=100" width="100px;" alt="Brennan Kinney"/><br /><sub><b>Brennan Kinney</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=polarathene" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/LongerHV"><img src="https://avatars.githubusercontent.com/u/46924944?v=4?s=100" width="100px;" alt="Michał Mieszczak"/><br /><sub><b>Michał Mieszczak</b></sub></a><br /><a href="#ideas-LongerHV" title="Ideas, Planning, & Feedback">🤔</a> <a href="https://github.com/authelia/authelia/commits?author=LongerHV" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/paul-ohl"><img src="https://avatars.githubusercontent.com/u/37795294?v=4?s=100" width="100px;" alt="Paul Ohl"/><br /><sub><b>Paul Ohl</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=paul-ohl" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/smkent"><img src="https://avatars.githubusercontent.com/u/2831985?v=4?s=100" width="100px;" alt="Stephen Kent"/><br /><sub><b>Stephen Kent</b></sub></a><br /><a href="#ideas-smkent" title="Ideas, Planning, & Feedback">🤔</a> <a href="https://github.com/authelia/authelia/commits?author=smkent" title="Code">💻</a> <a href="#design-smkent" title="Design">🎨</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Ohelig"><img src="https://avatars.githubusercontent.com/u/5841980?v=4?s=100" width="100px;" alt="Ohelig"/><br /><sub><b>Ohelig</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Ohelig" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/chillinPanda"><img src="https://avatars.githubusercontent.com/u/250694?v=4?s=100" width="100px;" alt="Dinh Bao Dang"/><br /><sub><b>Dinh Bao Dang</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=chillinPanda" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/levkoburburas"><img src="https://avatars.githubusercontent.com/u/62853952?v=4?s=100" width="100px;" alt="levkoburburas"/><br /><sub><b>levkoburburas</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=levkoburburas" title="Code">💻</a> <a href="#ideas-levkoburburas" title="Ideas, Planning, & Feedback">🤔</a> <a href="https://github.com/authelia/authelia/issues?q=author%3Alevkoburburas" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tiuub"><img src="https://avatars.githubusercontent.com/u/46517077?v=4?s=100" width="100px;" alt="tiuub"/><br /><sub><b>tiuub</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=tiuub" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://joshgordon.net"><img src="https://avatars.githubusercontent.com/u/2125341?v=4?s=100" width="100px;" alt="Josh Gordon"/><br /><sub><b>Josh Gordon</b></sub></a><br /><a href="#ideas-joshgordon" title="Ideas, Planning, & Feedback">🤔</a> <a href="#security-joshgordon" title="Security">🛡️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/silasfrancisco"><img src="https://avatars.githubusercontent.com/u/84447762?v=4?s=100" width="100px;" alt="silasfrancisco"/><br /><sub><b>silasfrancisco</b></sub></a><br /><a href="#security-silasfrancisco" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/n4m3l3ss-b0t"><img src="https://avatars.githubusercontent.com/u/1162710?v=4?s=100" width="100px;" alt="Ricardo Pesqueira"/><br /><sub><b>Ricardo Pesqueira</b></sub></a><br /><a href="#security-n4m3l3ss-b0t" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/HaroldVB"><img src="https://avatars.githubusercontent.com/u/73724671?v=4?s=100" width="100px;" alt="Harold"/><br /><sub><b>Harold</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=HaroldVB" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Crowley723"><img src="https://avatars.githubusercontent.com/u/26265198?v=4?s=100" width="100px;" alt="Brynn Crowley"/><br /><sub><b>Brynn Crowley</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Crowley723" title="Documentation">📖</a> <a href="#design-Crowley723" title="Design">🎨</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://budimanjojo.com"><img src="https://avatars.githubusercontent.com/u/13085918?v=4?s=100" width="100px;" alt="Budiman Jojo"/><br /><sub><b>Budiman Jojo</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=budimanjojo" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/hendrik1120"><img src="https://avatars.githubusercontent.com/u/89412959?v=4?s=100" width="100px;" alt="Hendrik Sievers"/><br /><sub><b>Hendrik Sievers</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=hendrik1120" title="Documentation">📖</a> <a href="#design-hendrik1120" title="Design">🎨</a> <a href="#ideas-hendrik1120" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/m-georgi"><img src="https://avatars.githubusercontent.com/u/20987691?v=4?s=100" width="100px;" alt="Marcus Georgi"/><br /><sub><b>Marcus Georgi</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=m-georgi" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/samos667"><img src="https://avatars.githubusercontent.com/u/50653464?v=4?s=100" width="100px;" alt="samos667"/><br /><sub><b>samos667</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=samos667" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/0xSysR3ll"><img src="https://avatars.githubusercontent.com/u/31414959?v=4?s=100" width="100px;" alt="0xsysr3ll"/><br /><sub><b>0xsysr3ll</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=0xSysR3ll" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/cromelex"><img src="https://avatars.githubusercontent.com/u/96779452?v=4?s=100" width="100px;" alt="Dan"/><br /><sub><b>Dan</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=cromelex" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://shaamallow.com"><img src="https://avatars.githubusercontent.com/u/39766320?v=4?s=100" width="100px;" alt="Eyal Benaroche"/><br /><sub><b>Eyal Benaroche</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Shaamallow" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/wangweixuan"><img src="https://avatars.githubusercontent.com/u/24620923?v=4?s=100" width="100px;" alt="Wang Weixuan"/><br /><sub><b>Wang Weixuan</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Awangweixuan" title="Bug reports">🐛</a> <a href="https://github.com/authelia/authelia/commits?author=wangweixuan" title="Code">💻</a> <a href="https://github.com/authelia/authelia/commits?author=wangweixuan" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://t.me/daniw1337"><img src="https://avatars.githubusercontent.com/u/21097466?v=4?s=100" width="100px;" alt="Dani"/><br /><sub><b>Dani</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=DaniW42" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://lhns.de"><img src="https://avatars.githubusercontent.com/u/1524059?v=4?s=100" width="100px;" alt="Pierre Kisters"/><br /><sub><b>Pierre Kisters</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=lhns" title="Code">💻</a> <a href="https://github.com/authelia/authelia/issues?q=author%3Alhns" title="Bug reports">🐛</a> <a href="https://github.com/authelia/authelia/commits?author=lhns" title="Tests">⚠️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="http://auston.dev"><img src="https://avatars.githubusercontent.com/u/5049050?v=4?s=100" width="100px;" alt="Auston Pramodh Barboza"/><br /><sub><b>Auston Pramodh Barboza</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=austonpramodh" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ThomasSteinbach"><img src="https://avatars.githubusercontent.com/u/1683246?v=4?s=100" width="100px;" alt="Thomas Steinbach"/><br /><sub><b>Thomas Steinbach</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=ThomasSteinbach" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Steve-Brule"><img src="https://avatars.githubusercontent.com/u/83868338?v=4?s=100" width="100px;" alt="Steve-Brule"/><br /><sub><b>Steve-Brule</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Steve-Brule" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/peter-"><img src="https://avatars.githubusercontent.com/u/4047365?v=4?s=100" width="100px;" alt="peter"/><br /><sub><b>peter</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Apeter-" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://ocnr.org"><img src="https://avatars.githubusercontent.com/u/1495157?v=4?s=100" width="100px;" alt="Nick O'Connor"/><br /><sub><b>Nick O'Connor</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Anick-oconnor" title="Bug reports">🐛</a> <a href="#userTesting-nick-oconnor" title="User Testing">📓</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/pedorich-n"><img src="https://avatars.githubusercontent.com/u/15573098?v=4?s=100" width="100px;" alt="Nikita Pedorich"/><br /><sub><b>Nikita Pedorich</b></sub></a><br /><a href="#userTesting-pedorich-n" title="User Testing">📓</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/atoerien"><img src="https://avatars.githubusercontent.com/u/49614525?v=4?s=100" width="100px;" alt="Andre Toerien"/><br /><sub><b>Andre Toerien</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Aatoerien" title="Bug reports">🐛</a> <a href="https://github.com/authelia/authelia/commits?author=atoerien" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/erusc"><img src="https://avatars.githubusercontent.com/u/208130995?v=4?s=100" width="100px;" alt="erusc"/><br /><sub><b>erusc</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Aerusc" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://kevincox.ca"><img src="https://avatars.githubusercontent.com/u/494012?v=4?s=100" width="100px;" alt="Kevin Cox"/><br /><sub><b>Kevin Cox</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Akevincox" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tkf144"><img src="https://avatars.githubusercontent.com/u/17581159?v=4?s=100" width="100px;" alt="Tom"/><br /><sub><b>Tom</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=tkf144" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://shiziblog.cn"><img src="https://avatars.githubusercontent.com/u/29810238?v=4?s=100" width="100px;" alt="Br1an"/><br /><sub><b>Br1an</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=Br1an67" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/andreasbrett"><img src="https://avatars.githubusercontent.com/u/6610451?v=4?s=100" width="100px;" alt="Andreas Brett"/><br /><sub><b>Andreas Brett</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=andreasbrett" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/FlorianObermayer"><img src="https://avatars.githubusercontent.com/u/4497761?v=4?s=100" width="100px;" alt="FO"/><br /><sub><b>FO</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=FlorianObermayer" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/x0k"><img src="https://avatars.githubusercontent.com/u/23502797?v=4?s=100" width="100px;" alt="Roman Krasilnikov"/><br /><sub><b>Roman Krasilnikov</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=x0k" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tomaszduda23"><img src="https://avatars.githubusercontent.com/u/35012788?v=4?s=100" width="100px;" alt="tomaszduda23"/><br /><sub><b>tomaszduda23</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=tomaszduda23" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://louisonsarlinmagnus.github.io/"><img src="https://avatars.githubusercontent.com/u/70757160?v=4?s=100" width="100px;" alt="Louison SARLIN--MAGNUS"/><br /><sub><b>Louison SARLIN--MAGNUS</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=louisonsarlinmagnus" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/AlexViridi"><img src="https://avatars.githubusercontent.com/u/80551059?v=4?s=100" width="100px;" alt="AlexViridi"/><br /><sub><b>AlexViridi</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=AlexViridi" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/xaabi6"><img src="https://avatars.githubusercontent.com/u/22637294?v=4?s=100" width="100px;" alt="Xabi"/><br /><sub><b>Xabi</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=xaabi6" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://caiocdcs.github.io"><img src="https://avatars.githubusercontent.com/u/8338516?v=4?s=100" width="100px;" alt="Caio Silva"/><br /><sub><b>Caio Silva</b></sub></a><br /><a href="https://github.com/authelia/authelia/commits?author=caiocdcs" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/deprrous"><img src="https://avatars.githubusercontent.com/u/144512455?v=4?s=100" width="100px;" alt="deprrous"/><br /><sub><b>deprrous</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Adeprrous" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/distorx"><img src="https://avatars.githubusercontent.com/u/1065150?v=4?s=100" width="100px;" alt="Ricardo Rivera"/><br /><sub><b>Ricardo Rivera</b></sub></a><br /><a href="https://github.com/authelia/authelia/issues?q=author%3Adistorx" title="Bug reports">🐛</a> <a href="https://github.com/authelia/authelia/commits?author=distorx" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification.
Contributions of any kind welcome!

### Sponsors

***Help Wanted:*** We are actively looking for sponsorship to obtain either a code security audit, penetration testing,
or other audits related to improving the security of Authelia.

Any company can become a sponsor by donating or providing any benefit to the project or the team helping improve
Authelia.

#### JetBrains

Thank you to [<img src="https://www.authelia.com/svgs/logos/jetbrains.svg" alt="JetBrains" width="32"> JetBrains](https://www.jetbrains.com/?from=Authelia)
for providing us with free licenses to their great tools.

* [<img src="https://www.authelia.com/svgs/logos/intellij-idea.svg" alt="IDEA" width="32"> IDEA](http://www.jetbrains.com/idea/)
* [<img src="https://www.authelia.com/svgs/logos/goland.svg" alt="GoLand" width="32"> GoLand](http://www.jetbrains.com/go/)
* [<img src="https://www.authelia.com/svgs/logos/webstorm.svg" alt="WebStorm" width="32"> WebStorm](http://www.jetbrains.com/webstorm/)

#### Microsoft

Our pipeline agents which we rely on for productivity are hosted on [Azure](https://azure.microsoft.com/?from=Authelia)
and our [git repositories](https://github.com/authelia) are hosted on [GitHub](https://github.com/?from=Authela)
which are both [Microsoft](https://www.microsoft.com/?from=Authelia) products.

[<img src="https://www.authelia.com/svgs/logos/microsoft.svg" alt="microsoft" height="32">](https://www.microsoft.com/?from=Authelia)

[<img src="https://www.authelia.com/svgs/logos/azure.svg" alt="Azure" height="32">](https://azure.microsoft.com/?from=Authelia)

### Open Collective

#### Backers

Thank you to all our backers! 🙏 [Become a backer](https://opencollective.com/authelia-sponsors/contribute) and help us
sustain our community. The money we currently receive is dedicated to fund a security audit, and potentially in the
future introducing a bug bounty program to give us as many
eyes as we can to detect potential vulnerabilities.
<a href="https://opencollective.com/authelia-sponsors#backers"><img src="https://opencollective.com/authelia-sponsors/backers.svg?width=890"></a>

#### Sponsorship

Companies contributing to Authelia via Open Collective will have a special mention below.
[Become a sponsor](https://opencollective.com/authelia-sponsors#sponsor).

<a href="https://opencollective.com/authelia-sponsors/sponsor/0/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/0/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/1/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/1/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/2/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/2/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/3/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/3/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/4/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/4/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/5/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/5/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/6/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/6/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/7/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/7/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/8/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/8/avatar.svg"></a>
<a href="https://opencollective.com/authelia-sponsors/sponsor/9/website"><img src="https://opencollective.com/authelia-sponsors/sponsor/9/avatar.svg"></a>

## License

**Authelia** is **licensed** under the **[Apache 2.0]** license. The terms of the license are detailed in
[LICENSE](LICENSE).

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fauthelia%2Fauthelia.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fauthelia%2Fauthelia?ref=badge_large)


[Apache 2.0]: https://www.apache.org/licenses/LICENSE-2.0
[TOTP]: https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm
[FIDO2]: https://www.yubico.com/authentication-standards/fido2/
[YubiKey]: https://www.yubico.com/products/yubikey-5-overview/
[WebAuthn]: https://www.yubico.com/authentication-standards/webauthn/
[auth_request]: https://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[config.template.yml]: ./config.template.yml
[nginx]: https://www.authelia.com/integration/proxies/nginx/
[Traefik]: https://www.authelia.com/integration/proxies/traefik/
[Caddy]: https://www.authelia.com/integration/proxies/caddy/
[Skipper]: https://www.authelia.com/integration/proxies/skipper/
[Envoy]: https://www.authelia.com/integration/proxies/envoy/
[HAProxy]: https://www.authelia.com/integration/proxies/haproxy/
[Docker]: https://docker.com/
[Kubernetes]: https://kubernetes.io/
[security]: https://github.com/authelia/authelia/security/policy
[#support]: https://discord.com/channels/707844280412012608/707844280412012612
[#contributing]: https://discord.com/channels/707844280412012608/804943261265297408
[OpenID Certified™]: https://openid.net/certification/
[OpenID Connect™ protocol]: https://openid.net/developers/how-connect-works/
