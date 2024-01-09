
1) To run local dev:
    ```bash
    docker run --rm \
          --volume="$PWD:/srv/jekyll:Z" \
          --publish [::1]:4000:4000   \
          jvconseil/jekyll-docker:4.3.2 \
          jekyll serve
    ```
    (please note that official jekyll image is not used because of the problems with just-the-docs theme).

2) Then open `http://localhost:4000/kube-arangodb/`

Note: if you change _config.yml, this command should be restarted to take effect.




Links:
- https://just-the-docs.com/docs/
- https://jekyllrb.com/docs/collections/#documents
