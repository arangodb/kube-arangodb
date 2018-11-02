#!/bin/bash
set -u

version="arangodb-preview:3.4.0-rc.3"
enterprise_secret="$ARANGO_EP_SECRET" #only the number
community="arangodb/$version"
enterprise="registry.arangodb.com/arangodb/$version-$enterprise_secret"
community="neunhoef/arangodb:3.4"
enterprise="neunhoef/arangodb:3.4"

rm -fr generated
mkdir -p generated

for path in *.template.yaml; do
    base_file="${path%.template.yaml}"
    target="./generated/$base_file-community-dev.yaml"
    cp "$path" "$target"
    sed -i "s|@IMAGE@|$community|" "$target"
    sed -i "s|@ENVIRONMENT@|Development|" "$target"
    echo "created $target"
done

for path in *.template.yaml; do
    base_file="${path%.template.yaml}"
    target="./generated/$base_file-community-pro.yaml"
    cp "$path" "$target"
    sed -i "s|@IMAGE@|$community|" "$target"
    sed -i "s|@ENVIRONMENT@|Production|" "$target"
    echo "created $target"
done

for path in *.template.yaml; do
    base_file="${path%.template.yaml}"
    target="./generated/$base_file-enterprise-dev.yaml"
    cp "$path" "$target"
    sed -i "s|@IMAGE@|$enterprise|" "$target"
    sed -i "s|@ENVIRONMENT@|Development|" "$target"
    echo "created $target"
done

for path in *.template.yaml; do
    base_file="${path%.template.yaml}"
    target="./generated/$base_file-enterprise-pro.yaml"
    cp "$path" "$target"
    sed -i "s|@IMAGE@|$enterprise|" "$target"
    sed -i "s|@ENVIRONMENT@|Production|" "$target"
    echo "created $target"
done
