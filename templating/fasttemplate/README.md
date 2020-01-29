fasttemplate
============

Simple and fast template engine for Go.
Forked from [fasttemplate](https://github.com/valyala/fasttemplate).

This package was modified from the original one:
1. usage of unsafe is removed
2. usage of buffer pools is removed

*Please note that fasttemplate doesn't do any escaping on template values
unlike [html/template](http://golang.org/pkg/html/template/) do. So values
must be properly escaped before passing them to fasttemplate.*

