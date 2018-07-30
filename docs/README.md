# sqlr documentation

This directory contains the restructured text source for the
detailed package documentation. This documentation is built
using the [Sphinx Document Generator](http://sphinx-doc.org).

You can view the generated documentation at https://jjeffery.github.io/sqlr.

To build the documentation:

1.  Clone the sqlr repository to an arbitrary location. The example uses `~/docs`.
    If you are using your own fork you will have to adjust the location in the `git clone`
    commannd.

```bash
mkdir -p ~/docs
cd ~/docs
git clone git@github.com:jjeffery/sqlr
cd sqlr
git checkout gh-pages
git symbolic-ref HEAD refs/heads/gh-pages  # auto-switches branches to gh-pages
```

2.  Confirm that you are are on the gh-pages branch.

```bash
git branch
```

3.  Create an environment variable `SQLR_GHPAGES` that points to the directory
    you have just created. The Makefile uses this environment variable to find
    the location to put the generated HTML.

```bash
echo "export SQLR_GHPAGES=$PWD" >> ~/.bash_profile
. ~/.bash_profile
echo $SQLR_GHPAGES
```

4.  Change directory to the docs source directory and run the document generator.

```bash
cd $GOPATH/src/github.com/jjeffery/sqlr/docs
make html
```

5.  The generated documentation is now present in the `~/docs/sqlr` directory.
    Commit to git and push, and the documentation is ready to view.

```bash
cd ~/docs/sqlr
git add -A .
git commit -m "update documentation"
git push
```
