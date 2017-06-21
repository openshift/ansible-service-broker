# Building

```
dnf install tito
```

## tagging a release

Tito tag will create a tag for the project using the project name and the
version-release from the spec file. This creates a nice consistent tagging
stream.

```
tito tag
# follow the prompts.
# inspect the changelog to ensure it looks acceptable
# save file
git push origin
git push origin NAMEOFTAG
tito release brew
```

## building rpm locally

To build an rpm of the currently committed (but not necessarily pushed
to origin) you can use the `--test` flag.

```
tito build --test --rpm
```

## building rpm from last tag

To build an rpm of the last pushed tag, simply run:

```
tito build --rpm
```
