default:
    just -l

test-build:
    nix build --no-link .#
