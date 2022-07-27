#!/bin/bash
ls *.md | while read f; do C=$(basename $f .md|tr [:upper:] [:lower:]); echo "- [$C](./$f)"; done >> INDEX.md
