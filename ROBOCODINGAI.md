# OpenFox Platform



This is the OpenFox product fork of [New API](https://github.com/QuantumNous/new-api). Repository: `qy-robot/robocodingai-platform`; product branch: `robo/main`. The original upstream `main` branch is preserved.

## Component ownership

- `web/`: the existing website/admin frontend.
- `router/`, `middleware/`, `controller/`, `service/`, `model/`, `relay/`: the existing Go backend, account, wallet and model relay.
- The separate OpenFox desktop uses DSH Desktop; the upstream `electron/` directory is not our chosen App implementation.
- Company proprietary knowledge belongs to the separate private capability service and is not stored here.

`robocodingai.product.json` records product naming only. Desktop authorization, branding integration, shared-wallet behavior and deployment are not implemented by repository initialization. The inspected source baseline remains fixed for later integration work.

Preserve the original module/package paths, frontend build, licenses, source attribution and UI notice requirements. This fork remains subject to the original [LICENSE](LICENSE) and [README licensing terms](README.md). Product naming does not replace those obligations.
