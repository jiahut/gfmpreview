gfmpreview
==========

This is a small tool that renders your local GitHub Flavored Markdown files.

It works by scanning the directory where it's run from so you don't have to do anything.

Installing
----------

If you have Go installed:

    go get github.com/vrischmann/gfmpreview

Otherwise look at the [releases](https://github.com/vrischmann/gfmpreview/releases) page.

Running
-------

Simply run `gfmpreview` in the directory where your md files are located. It can also be any parent directory since *gfmpreview* scans for markdown files recursively.

If the default listening port of *3030* does not work for you you have the option to change it:

    gfmpreview -l ":4040"

Or you can change the listening address completely:

    gfmpreview -l "10.0.0.1:4040"


If you don't pass a path in the URL *gfmpreview* will render a list of all markdown files relative to the current working directory:

![Listing view](https://vrischmann.me/files/gfmpreview/listing.png)

If you pass a correct path relative to the current working directory *gfmpreview* will do the actual rendering:

![Preview](https://vrischmann.me/files/gfmpreview/preview.png)

License
-------

MIT. See the LICENSE file.
