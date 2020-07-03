while sleep 1; do
  find . \
    \( -iname '*.go' -o -iname '.html' -o -iname '*.yml' -o -iname '*.json' \) \
    -a -not -path './tmp/*' \
    -a -not -path './vendor/*' \
    -a -not -iname '*_test.go' \
    | entr -dr go run *.go
done
