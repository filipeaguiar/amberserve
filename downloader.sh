#!/bin/bash
REPO=/roms
IA=/roms/ports/amberserver/ia

$IA download --no-directories "$1" --glob"$2" --destdir "$REPO/"