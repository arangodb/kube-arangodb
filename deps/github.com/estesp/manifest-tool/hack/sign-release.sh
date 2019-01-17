#!/bin/bash

for file in manifest-tool-*; do
	echo "Signing ${file}.."
	gpg --armor --detach-sign "${file}"
done

# quick verify step
for file in manifest-tool-*.asc; do
	echo "Verifying ${file}.."
	gpg --verify "${file}" 1>/dev/null 2>&1 || echo "Verify signature for ${file} failed!"
done
