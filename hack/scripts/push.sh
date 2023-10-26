#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

$SCRIPTS/image.sh

buildah push "${IMAGE}"
