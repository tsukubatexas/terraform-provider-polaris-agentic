## Summary

-

## Checks

- [ ] `make generate fmt test build`
- [ ] `bash -n scripts/*.sh`
- [ ] `scripts/check_release_please_config.sh`
- [ ] `scripts/check_actions_pinned.sh`
- [ ] ADR added or updated when behavior/policy changed
- [ ] Real Polaris gate run when provider/runtime behavior changed

## Security

- [ ] No secrets, tokens, Terraform state, or cloud identifiers are committed
- [ ] Workflow permissions remain least-privilege
- [ ] External GitHub Actions remain pinned to full commit SHAs
